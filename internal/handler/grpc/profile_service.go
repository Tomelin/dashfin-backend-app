package grpc

import (
	"context"
	"reflect" // Required for UpdateProfile if using reflection, or manual checks

	"cloud.google.com/go/firestore"
	"example.com/profile-service/internal/auth" // For UserIDFromContext
	"example.com/profile-service/internal/database"
	"example.com/profile-service/internal/domain"
	pb "example.com/profile-service/pkg/grpc/proto" // Alias for generated proto
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb" // For DeleteProfileResponse
)

// ProfileGrpcServer implements the proto.ProfileServiceServer interface.
type ProfileGrpcServer struct {
	pb.UnimplementedProfileServiceServer // Recommended for forward compatibility
	FirestoreClient                      *firestore.Client
}

// NewProfileGrpcServer creates a new ProfileGrpcServer.
func NewProfileGrpcServer(client *firestore.Client) *ProfileGrpcServer {
	return &ProfileGrpcServer{FirestoreClient: client}
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

func (s *ProfileGrpcServer) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.CreateProfileResponse, error) {
	authUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || authUserID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user ID not found in context")
	}

	// Basic validation for required fields in the request
	if req.GetFullName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "full_name is required")
	}
	if req.GetEmail() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}
	// Add other validations as per domain.Profile tags if desired (e.g. email format)

	profileDomain := fromPbCreateRequestToDomain(req)

	err := database.CreateProfile(ctx, s.FirestoreClient, authUserID, profileDomain)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create profile: %v", err)
	}

	return &pb.CreateProfileResponse{
		Profile: toPbProfile(profileDomain, authUserID),
	}, nil
}

func (s *ProfileGrpcServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	requestUserID := req.GetUserId()
	if requestUserID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id in request is required")
	}

	authUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || authUserID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user ID not found in context")
	}

	if authUserID != requestUserID {
		return nil, status.Errorf(codes.PermissionDenied, "you are not authorized to access this profile")
	}

	profileDomain, err := database.GetProfile(ctx, s.FirestoreClient, requestUserID)
	if err != nil {
		dbStatus, _ := status.FromError(err) // Convert to status.Status to check code
		if dbStatus.Code() == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "profile not found for user_id: %s", requestUserID)
		}
		return nil, status.Errorf(codes.Internal, "failed to get profile: %v", err)
	}

	return &pb.GetProfileResponse{
		Profile: toPbProfile(profileDomain, requestUserID),
	}, nil
}

// UpdateProfile implements the gRPC service method for updating a profile.
func (s *ProfileGrpcServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	requestUserID := req.GetUserId()
	if requestUserID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id in request is required")
	}

	authUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || authUserID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user ID not found in context")
	}

	if authUserID != requestUserID {
		return nil, status.Errorf(codes.PermissionDenied, "you are not authorized to update this profile")
	}

	updateMap := make(map[string]interface{})

	// For optional fields in proto, GetFieldName() returns the zero value if not set.
	// The .proto `optional string full_name` generates a `FullName *string` field in Go.
	// So req.FullName will be nil if not provided, or a pointer to a string if provided.
	// req.GetFullName() will dereference this or return "" if nil.
	// To check for explicit presence, we check the pointer.
	if req.FullName != nil {
		updateMap["fullName"] = req.GetFullName() // Or *req.FullName if you want to be explicit
	}
	if req.Email != nil { // Assuming email can be updated and is optional in Update request
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

	if len(updateMap) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no update fields provided")
	}

	err := database.UpdateProfile(ctx, s.FirestoreClient, requestUserID, updateMap)
	if err != nil {
		dbStatus, _ := status.FromError(err)
		if dbStatus.Code() == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "profile not found for user_id: %s to update", requestUserID)
		}
		if dbStatus.Code() == codes.InvalidArgument { // e.g. from database layer if map was empty after all
             return nil, status.Errorf(codes.InvalidArgument, "update arguments invalid: %v", err)
        }
		return nil, status.Errorf(codes.Internal, "failed to update profile: %v", err)
	}

	updatedProfileDomain, err := database.GetProfile(ctx, s.FirestoreClient, requestUserID)
	if err != nil {
		// Log that we updated but couldn't fetch, but still return success for the update itself
		// Or, decide if this is critical. For now, consider update successful if previous step had no error.
		return nil, status.Errorf(codes.Internal, "profile updated, but failed to retrieve updated version: %v", err)
	}

	return &pb.UpdateProfileResponse{
		Profile: toPbProfile(updatedProfileDomain, requestUserID),
	}, nil
}

// DeleteProfile implements the gRPC service method for deleting a profile.
func (s *ProfileGrpcServer) DeleteProfile(ctx context.Context, req *pb.DeleteProfileRequest) (*emptypb.Empty, error) {
	requestUserID := req.GetUserId()
	if requestUserID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id in request is required")
	}

	authUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || authUserID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user ID not found in context")
	}

	if authUserID != requestUserID {
		return nil, status.Errorf(codes.PermissionDenied, "you are not authorized to delete this profile")
	}

	err := database.DeleteProfile(ctx, s.FirestoreClient, requestUserID)
	if err != nil {
		// Firestore Delete often doesn't error on Not Found. Check for other errors.
		return nil, status.Errorf(codes.Internal, "failed to delete profile: %v", err)
	}

	return &emptypb.Empty{}, nil
}
