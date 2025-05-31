package grpc

import (
	"context"
	"reflect" // Required for UpdateProfile if using reflection, or manual checks
	"time"    // For timing requests

	"cloud.google.com/go/firestore"
	"example.com/profile-service/internal/auth" // For UserIDFromContext
	"example.com/profile-service/internal/database"
	"example.com/profile-service/internal/domain"
	pb "example.com/profile-service/pkg/grpc/proto" // Alias for generated proto
	otelCodes "go.opentelemetry.io/otel/codes"      // For span status
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb" // For DeleteProfileResponse

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const grpcHandlerMeterName = "example.com/profile-service/grpc-handler"
const ProfileServiceFullName = "/profile.ProfileService/" // Base for method names

// ProfileGrpcServer implements the proto.ProfileServiceServer interface.
type ProfileGrpcServer struct {
	pb.UnimplementedProfileServiceServer // Recommended for forward compatibility
	FirestoreClient             *firestore.Client
	RpcRequestsTotalCounter     metric.Int64Counter
	RpcRequestDurationSeconds metric.Float64Histogram
}

// NewProfileGrpcServer creates a new ProfileGrpcServer.
func NewProfileGrpcServer(client *firestore.Client) *ProfileGrpcServer {
	meter := otel.Meter(grpcHandlerMeterName)
	requestsCounter, rcErr := meter.Int64Counter("rpc.server.requests_total",
		metric.WithDescription("Total number of RPC requests."),
		metric.WithUnit("{request}"),
	)
	durationHistogram, rhErr := meter.Float64Histogram("rpc.server.duration_seconds",
		metric.WithDescription("RPC request duration in seconds."),
		metric.WithUnit("s"),
	)
	if rcErr != nil {
		otel.Handle(rcErr)
	}
	if rhErr != nil {
		otel.Handle(rhErr)
	}

	return &ProfileGrpcServer{
		FirestoreClient:             client,
		RpcRequestsTotalCounter:     requestsCounter,
		RpcRequestDurationSeconds: durationHistogram,
	}
}

// Helper function to convert domain.Profile to pb.Profile
func toPbProfile(dProfile *domain.Profile, userID string) *pb.Profile {
	if dProfile == nil {
		return nil
	}
	// For pb.Profile, optional fields are pointers. We need to pass the address of the domain fields.
	// If domain fields can be empty and that's distinct from not being set,
	// domain struct might also need pointers or sql.NullString etc.
	// Assuming domain fields are direct values for now.
	pbProfile := &pb.Profile{
		UserId:    userID,
		FullName:  dProfile.FullName,
		Email:     dProfile.Email,
		// Optional fields in .proto (like optional string phone) generate as *string in Go.
		// So, we need to provide a pointer if the value is not its zero value.
		// If the domain field is an empty string, we might want to pass nil or &"" based on semantics.
		// For simplicity, if domain field is not empty, pass its address. Otherwise, pass nil.
	}
	if dProfile.Phone != "" {
		pbProfile.Phone = &dProfile.Phone
	}
	if dProfile.BirthDate != "" {
		pbProfile.BirthDate = &dProfile.BirthDate
	}
	if dProfile.CEP != "" {
		pbProfile.Cep = &dProfile.CEP
	}
	if dProfile.City != "" {
		pbProfile.City = &dProfile.City
	}
	if dProfile.State != "" {
		pbProfile.State = &dProfile.State
	}
	return pbProfile
}

// Helper function to convert pb.CreateProfileRequest to domain.Profile
func fromPbCreateRequestToDomain(req *pb.CreateProfileRequest) *domain.Profile {
	return &domain.Profile{
		FullName:  req.FullName, // required string
		Email:     req.Email,    // required string
		Phone:     req.GetPhone(), // GetPhone() handles nil for optional fields, returns zero value if not set
		BirthDate: req.GetBirthDate(),
		CEP:       req.GetCep(),
		City:      req.GetCity(),
		State:     req.GetState(),
	}
}

func (s *ProfileGrpcServer) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (res *pb.CreateProfileResponse, err error) {
	startTime := time.Now()
	span := trace.SpanFromContext(ctx) // Get span from context (populated by interceptor)
	span.AddEvent("Handling CreateProfile RPC")
	methodName := ProfileServiceFullName + "CreateProfile"

	defer func() {
		statusCode := status.Code(err)
		commonAttrs := []attribute.KeyValue{
			attribute.String("rpc.method", methodName),
			attribute.String("rpc.system", "grpc"),
			attribute.Int("rpc.grpc.status_code", int(statusCode)),
		}
		s.RpcRequestsTotalCounter.Add(ctx, 1, metric.WithAttributes(commonAttrs...))
		s.RpcRequestDurationSeconds.Record(ctx, time.Since(startTime).Seconds(), metric.WithAttributes(commonAttrs...))
		if err != nil && span.IsRecording() {
			span.RecordError(err)
			// Interceptor will set overall status, but we can add more detail here if needed.
			// For example, if a specific sub-operation failed that isn't the final gRPC status.
			// span.SetStatus(otelCodes.Error, err.Error())
		}
	}()

	authUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || authUserID == "" {
		err = status.Errorf(codes.Unauthenticated, "user ID not found in context")
		return nil, err
	}
	span.SetAttributes(attribute.String("enduser.id_from_context", authUserID))

	if req.GetFullName() == "" {
		err = status.Errorf(codes.InvalidArgument, "full_name is required")
		return nil, err
	}
	if req.GetEmail() == "" {
		err = status.Errorf(codes.InvalidArgument, "email is required")
		return nil, err
	}
	span.AddEvent("Request validated")

	profileDomain := fromPbCreateRequestToDomain(req)

	dbErr := database.CreateProfile(ctx, s.FirestoreClient, authUserID, profileDomain)
	if dbErr != nil {
		err = status.Errorf(codes.Internal, "failed to create profile: %v", dbErr)
		return nil, err
	}
	span.AddEvent("Profile created in database")

	return &pb.CreateProfileResponse{
		Profile: toPbProfile(profileDomain, authUserID),
	}, nil
}

func (s *ProfileGrpcServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (res *pb.GetProfileResponse, err error) {
	startTime := time.Now()
	span := trace.SpanFromContext(ctx)
	span.AddEvent("Handling GetProfile RPC")
	methodName := ProfileServiceFullName + "GetProfile"

	defer func() {
		statusCode := status.Code(err)
		commonAttrs := []attribute.KeyValue{
			attribute.String("rpc.method", methodName),
			attribute.String("rpc.system", "grpc"),
			attribute.Int("rpc.grpc.status_code", int(statusCode)),
		}
		s.RpcRequestsTotalCounter.Add(ctx, 1, metric.WithAttributes(commonAttrs...))
		s.RpcRequestDurationSeconds.Record(ctx, time.Since(startTime).Seconds(), metric.WithAttributes(commonAttrs...))
		if err != nil && span.IsRecording() {
			span.RecordError(err)
		}
	}()

	requestUserID := req.GetUserId()
	if requestUserID == "" {
		err = status.Errorf(codes.InvalidArgument, "user_id in request is required")
		return nil, err
	}
	span.SetAttributes(attribute.String("profile.user_id_requested", requestUserID))

	authUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || authUserID == "" {
		err = status.Errorf(codes.Unauthenticated, "user ID not found in context")
		return nil, err
	}
	span.SetAttributes(attribute.String("enduser.id_from_context", authUserID))

	if authUserID != requestUserID {
		err = status.Errorf(codes.PermissionDenied, "you are not authorized to access this profile")
		return nil, err
	}

	profileDomain, dbErr := database.GetProfile(ctx, s.FirestoreClient, requestUserID)
	if dbErr != nil {
		dbStatus, _ := status.FromError(dbErr)
		if dbStatus.Code() == codes.NotFound {
			err = status.Errorf(codes.NotFound, "profile not found for user_id: %s", requestUserID)
			return nil, err
		}
		err = status.Errorf(codes.Internal, "failed to get profile: %v", dbErr)
		return nil, err
	}
	span.AddEvent("Profile retrieved from database")

	return &pb.GetProfileResponse{
		Profile: toPbProfile(profileDomain, requestUserID),
	}, nil
}

// UpdateProfile implements the gRPC service method for updating a profile.
func (s *ProfileGrpcServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (res *pb.UpdateProfileResponse, err error) {
	startTime := time.Now()
	span := trace.SpanFromContext(ctx)
	span.AddEvent("Handling UpdateProfile RPC")
	methodName := ProfileServiceFullName + "UpdateProfile"

	defer func() {
		statusCode := status.Code(err)
		commonAttrs := []attribute.KeyValue{
			attribute.String("rpc.method", methodName),
			attribute.String("rpc.system", "grpc"),
			attribute.Int("rpc.grpc.status_code", int(statusCode)),
		}
		s.RpcRequestsTotalCounter.Add(ctx, 1, metric.WithAttributes(commonAttrs...))
		s.RpcRequestDurationSeconds.Record(ctx, time.Since(startTime).Seconds(), metric.WithAttributes(commonAttrs...))
		if err != nil && span.IsRecording() {
			span.RecordError(err)
		}
	}()

	requestUserID := req.GetUserId()
	if requestUserID == "" {
		err = status.Errorf(codes.InvalidArgument, "user_id in request is required")
		return nil, err
	}
	span.SetAttributes(attribute.String("profile.user_id_requested", requestUserID))

	authUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || authUserID == "" {
		err = status.Errorf(codes.Unauthenticated, "user ID not found in context")
		return nil, err
	}
	span.SetAttributes(attribute.String("enduser.id_from_context", authUserID))

	if authUserID != requestUserID {
		err = status.Errorf(codes.PermissionDenied, "you are not authorized to update this profile")
		return nil, err
	}

	updateMap := make(map[string]interface{})

	if req.FullName != nil {
		updateMap["fullName"] = req.GetFullName()
	}
	if req.Email != nil {
		updateMap["email"] = req.GetEmail()
	}
	if req.Phone != nil {
		updateMap["phone"] = req.GetPhone()
	}
	if req.BirthDate != nil {
		updateMap["birthDate"] = req.GetBirthDate()
	}
	if req.Cep != nil {
		updateMap["cep"] = req.GetCep()
	}
	if req.City != nil {
		updateMap["city"] = req.GetCity()
	}
	if req.State != nil {
		updateMap["state"] = req.GetState()
	}
	span.AddEvent("Update map created", trace.WithAttributes(attribute.Int("update_map.size", len(updateMap))))

	if len(updateMap) == 0 {
		err = status.Errorf(codes.InvalidArgument, "no update fields provided")
		return nil, err
	}

	dbErr := database.UpdateProfile(ctx, s.FirestoreClient, requestUserID, updateMap)
	if dbErr != nil {
		dbStatus, _ := status.FromError(dbErr)
		if dbStatus.Code() == codes.NotFound {
			err = status.Errorf(codes.NotFound, "profile not found for user_id: %s to update", requestUserID)
			return nil, err
		}
		if dbStatus.Code() == codes.InvalidArgument {
			err = status.Errorf(codes.InvalidArgument, "update arguments invalid: %v", dbErr)
			return nil, err
		}
		err = status.Errorf(codes.Internal, "failed to update profile: %v", dbErr)
		return nil, err
	}
	span.AddEvent("Profile updated in database")

	updatedProfileDomain, dbErr := database.GetProfile(ctx, s.FirestoreClient, requestUserID)
	if dbErr != nil {
		err = status.Errorf(codes.Internal, "profile updated, but failed to retrieve updated version: %v", dbErr)
		return nil, err
	}
	span.AddEvent("Updated profile retrieved")

	return &pb.UpdateProfileResponse{
		Profile: toPbProfile(updatedProfileDomain, requestUserID),
	}, nil
}

// DeleteProfile implements the gRPC service method for deleting a profile.
func (s *ProfileGrpcServer) DeleteProfile(ctx context.Context, req *pb.DeleteProfileRequest) (res *emptypb.Empty, err error) {
	startTime := time.Now()
	span := trace.SpanFromContext(ctx)
	span.AddEvent("Handling DeleteProfile RPC")
	methodName := ProfileServiceFullName + "DeleteProfile"

	defer func() {
		statusCode := status.Code(err)
		commonAttrs := []attribute.KeyValue{
			attribute.String("rpc.method", methodName),
			attribute.String("rpc.system", "grpc"),
			attribute.Int("rpc.grpc.status_code", int(statusCode)),
		}
		s.RpcRequestsTotalCounter.Add(ctx, 1, metric.WithAttributes(commonAttrs...))
		s.RpcRequestDurationSeconds.Record(ctx, time.Since(startTime).Seconds(), metric.WithAttributes(commonAttrs...))
		if err != nil && span.IsRecording() {
			span.RecordError(err)
		}
	}()

	requestUserID := req.GetUserId()
	if requestUserID == "" {
		err = status.Errorf(codes.InvalidArgument, "user_id in request is required")
		return nil, err
	}
	span.SetAttributes(attribute.String("profile.user_id_requested", requestUserID))

	authUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || authUserID == "" {
		err = status.Errorf(codes.Unauthenticated, "user ID not found in context")
		return nil, err
	}
	span.SetAttributes(attribute.String("enduser.id_from_context", authUserID))

	if authUserID != requestUserID {
		err = status.Errorf(codes.PermissionDenied, "you are not authorized to delete this profile")
		return nil, err
	}

	dbErr := database.DeleteProfile(ctx, s.FirestoreClient, requestUserID)
	if dbErr != nil {
		err = status.Errorf(codes.Internal, "failed to delete profile: %v", dbErr)
		return nil, err
	}
	span.AddEvent("Profile deleted from database")

	return &emptypb.Empty{}, nil
}
