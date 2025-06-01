package support

import (
	// "context"
	"errors" // Make sure this is uncommented
	// "log"

	"github.com/graphql-go/graphql"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/handler"
)

// GraphQLHandler holds dependencies for GraphQL, e.g., the service.
// For now, it's a simple struct to hold the schema.
type GraphQLHandler struct {
	Schema graphql.Schema
	// SupportService *Service // Would be used by resolvers
}

// NewGraphQLHandler initializes the GraphQL schema and returns a handler.
// func NewGraphQLHandler(service *Service) (*GraphQLHandler, error) {
// resolver := NewGraphQLResolver(service) // If using a resolver struct

func NewGraphQLHandler() (*GraphQLHandler, error) { // Simpler version without service for now
	mutationFields := graphql.Fields{
		"createSupportRequest": &graphql.Field{
			Type: createSupportResponseObjectType, // Custom response type
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(supportRequestInputType),
				},
				// Headers like X-APP, X-USERID, X-AUTHORIZATION need to be accessed
				// from the HTTP context within the resolver.
				// GraphQL-go/handler passes the http.Request via context.
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				// Extract input object
				inputMap, _ := params.Args["input"].(map[string]interface{})
				category, _ := inputMap["category"].(string)
				description, _ := inputMap["description"].(string)

				// Validate Category enum
				if !SupportRequestCategory(category).IsValid() {
					return nil, errors.New("Invalid category value: " + category)
				}

				// Validate description length (mirroring HTTP/gRPC)
				if len(description) < 10 || len(description) > 2000 {
					return nil, errors.New("Description length must be between 10 and 2000 characters")
				}

				// Access headers from context (passed by graphql-go/handler)
				// This part requires the Gin context to be propagated correctly or accessed.
				// The `graphql-go/handler` typically makes `http.Request` available.
				// We'll assume for now that the auth middleware (for Gin) has already run
				// if this GraphQL endpoint is also served via Gin and protected.

				// httpRequest, ok := params.Context.Value("httpResponseWriter").(http.ResponseWriter) // This is one way
				// ginContext, ok := params.Context.Value("ginContextKey").(*gin.Context) // Needs custom setup

				// For now, let's assume these would be retrieved from context after middleware.
				// firebaseUID := "graphql-user-placeholder" // Get from context
				// appName := "graphql-app-placeholder"   // Get from context

				// log.Printf("GraphQL: Received support request from UID: %s for app: %s", firebaseUID, appName)
				// log.Printf("GraphQL: Request details: Category: %s, Description: %s", category, description)

				// Placeholder for service call
				// submittedRequest := SupportRequest{Category: category, Description: description}
				// _, err := gqlHandler.SupportService.Create(params.Context, submittedRequest, firebaseUID, appName)
				// if err != nil {
				//    return nil, err
				// }

				return map[string]interface{}{
					"status_message": "Support request received successfully via GraphQL.",
					"request_id":     "gql-placeholder-id",
					"submitted_data": map[string]string{"category": category, "description": description},
				}, nil
			},
		},
	}
	rootMutation := graphql.NewObject(graphql.ObjectConfig{Name: "RootMutation", Fields: mutationFields})

	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		// Query:    rootQuery, // No queries defined for now
		Mutation: rootMutation,
	})
	if err != nil {
		// log.Fatalf("Failed to create GraphQL schema: %v", err)
		return nil, err
	}

	return &GraphQLHandler{Schema: Schema /*, SupportService: service */}, nil
}

// ServeGraphQL creates a Gin handler func for GraphQL requests.
// It assumes AuthMiddleware has already run on the route if authentication is needed.
func (h *GraphQLHandler) ServeGraphQL() gin.HandlerFunc {
	// Configure the graphql-go/handler
	gqlHandler := handler.New(&handler.Config{
		Schema:   &h.Schema,
		Pretty:   true,
		GraphiQL: true, // Enable GraphiQL interface for easy testing
		// Playground: true, // Alternative to GraphiQL
	})

	return func(c *gin.Context) {
		// If you need to pass Gin-specific context to resolvers,
		// you might need to wrap or adapt `gqlHandler`.
		// For example, using context.WithValue to pass c.
		// ctx := context.WithValue(c.Request.Context(), "ginContextKey", c)
		// gqlHandler.ContextHandler(ctx, c.Writer, c.Request)

		// The AuthMiddleware should have already validated X-AUTHORIZATION, X-APP, X-USERID.
		// We can retrieve firebase_uid and app_name from gin context if needed by resolvers,
		// though typical GraphQL resolver context is `params.Context`.
		// firebaseUID, _ := c.Get("firebase_uid")
		// appName, _ := c.Get("app_name")
		// You might pass these into params.Context if your resolvers expect them there.
		// For example: params.Context = context.WithValue(params.Context, "firebaseUIDKey", firebaseUID)

		gqlHandler.ServeHTTP(c.Writer, c.Request)
	}
}
