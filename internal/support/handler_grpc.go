package support

import (
	"context"
	"errors"
	// "log" // For logging service calls

	pb "github.com/user/supportservice/pkg/grpc/supportpb" // Adjust import path
	// "google.golang.org/grpc/codes" // For gRPC status codes
	// "google.golang.org/grpc/status" // For gRPC status errors
)

// GRPCServer implements the gRPC server for the SupportService.
// It would typically hold a reference to the support service.
type GRPCServer struct {
	pb.UnimplementedSupportServiceServer // Recommended for forward compatibility
	// SupportService *Service              // Example: To be added when service layer exists
}

// NewGRPCServer creates a new GRPCServer.
// func NewGRPCServer(service *Service) *GRPCServer {
//     return &GRPCServer{SupportService: service}
// }

// CreateSupportRequest is the gRPC method for creating a support request.
// For gRPC, authentication (X-AUTHORIZATION) and other headers (X-APP, X-USERID)
// are typically handled by gRPC interceptors. The interceptor would validate them
// and could pass necessary info (like UID, app_name) via the context.
func (s *GRPCServer) CreateSupportRequest(ctx context.Context, req *pb.CreateSupportRequestRequest) (*pb.CreateSupportRequestResponse, error) {
	if req.GetRequest() == nil {
		return nil, errors.New("request payload is missing") // status.Errorf(codes.InvalidArgument, "request payload is missing")
	}

	category := req.GetRequest().GetCategory()
	description := req.GetRequest().GetDescription()
	xApp := req.GetXApp()
	xUserID := req.GetXUserid() // This would be validated against UID from token in an interceptor

	// Validate Category enum
	if !SupportRequestCategory(category).IsValid() {
		// return nil, status.Errorf(codes.InvalidArgument, "Invalid category value: %s", category)
		return nil, errors.New("Invalid category value: " + category)
	}

	// Validate description length (example, mirroring HTTP)
	if len(description) < 10 || len(description) > 2000 {
		// return nil, status.Errorf(codes.InvalidArgument, "Description length must be between 10 and 2000 characters")
		return nil, errors.New("Description length must be between 10 and 2000 characters")
	}

	// In a real scenario, an interceptor would have already validated X-AUTHORIZATION (Firebase token)
	// and potentially X-APP and X-USERID.
	// The UID from the token and X-APP would be available from ctx.
	// For now, we're using the values passed in the message for X-APP and X-USERID.
	// log.Printf("gRPC: Received support request from UserID: %s for app: %s", xUserID, xApp)
	// log.Printf("gRPC: Request details: Category: %s, Description: %s", category, description)

	// Placeholder for calling the service layer:
	// requestModel := SupportRequest{Category: category, Description: description}
	// _, err := s.SupportService.Create(ctx, requestModel, xUserID, xApp) // xUserID here assumes it's the Firebase UID
	// if err != nil {
	//     // return nil, status.Errorf(codes.Internal, "Failed to create support request: %v", err)
	//      return nil, errors.New("Failed to create support request: " + err.Error())
	// }

	return &pb.CreateSupportRequestResponse{
		StatusMessage: "Support request received successfully via gRPC.",
		RequestId:     "grpc-placeholder-id", // Example ID
	}, nil
}
