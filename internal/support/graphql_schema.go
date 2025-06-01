package support

import (
	"github.com/graphql-go/graphql"
	// pb "github.com/user/supportservice/pkg/grpc/supportpb" // For input/output types if needed
)

// supportRequestType defines the GraphQL type for SupportRequest
var supportRequestType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "SupportRequest",
		Fields: graphql.Fields{
			"category":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"description": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	},
)

// supportRequestInputType defines the GraphQL input type for creating a SupportRequest
// This helps in structuring the mutation arguments.
var supportRequestInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "SupportRequestInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"category":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"description": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		},
	},
)

// This will be the response type for our mutation
var createSupportResponseObjectType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "CreateSupportResponse",
		Fields: graphql.Fields{
			"status_message": &graphql.Field{Type: graphql.String},
			"request_id":     &graphql.Field{Type: graphql.String}, // Placeholder
			"submitted_data": &graphql.Field{Type: supportRequestType},
		},
	},
)

// Resolver functions will be defined in handler_graphql.go or here.
// For now, we'll define a placeholder resolver structure.
// type GraphQLResolver struct {
//     SupportService *Service // Dependency for actual logic
// }

// func NewGraphQLResolver(service *Service) *GraphQLResolver {
//     return &GraphQLResolver{SupportService: service}
// }

// This is a placeholder for the actual mutation field
// Real implementation will be in handler_graphql.go to access resolver methods
/*
var rootMutation = graphql.NewObject(graphql.ObjectConfig{
    Name: "RootMutation",
    Fields: graphql.Fields{
        "createSupportRequest": &graphql.Field{
            // ... (definition will be in handler_graphql.go)
        },
    },
})
*/

// Global schema variable, to be initialized in handler_graphql.go
var Schema graphql.Schema
