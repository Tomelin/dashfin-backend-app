package graphql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"example.com/profile-service/internal/auth"
	"example.com/profile-service/internal/database"
	"example.com/profile-service/internal/domain" // Now we need the concrete domain.Profile

	"github.com/graphql-go/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes" // aliased to avoid conflict
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	gRPCCodes "google.golang.org/grpc/codes" // For database error comparison, aliased
	"google.golang.org/grpc/status"
)

const gqlTracerName = "example.com/profile-service/graphql"

// Resolver holds dependencies for GraphQL resolvers.
type Resolver struct {
	FirestoreClient            *firestore.Client
	GqlRequestsTotalCounter    metric.Int64Counter
	GqlRequestDurationSeconds  metric.Float64Histogram
}

// NewResolver creates a new Resolver.
func NewResolver(client *firestore.Client) *Resolver {
	meter := otel.Meter(gqlHandlerMeterName) // Use meterName defined in schema.go
	requestsCounter, rcErr := meter.Int64Counter("gql.server.requests_total",
		metric.WithDescription("Total number of GraphQL requests."),
		metric.WithUnit("{request}"),
	)
	durationHistogram, rhErr := meter.Float64Histogram("gql.server.duration_seconds",
		metric.WithDescription("GraphQL request duration in seconds."),
		metric.WithUnit("s"),
	)
	if rcErr != nil {
		otel.Handle(rcErr)
	}
	if rhErr != nil {
		otel.Handle(rhErr)
	}
	return &Resolver{
		FirestoreClient:            client,
		GqlRequestsTotalCounter:    requestsCounter,
		GqlRequestDurationSeconds:  durationHistogram,
	}
}

// mapDomainProfileToGraphQL correctly converts domain.Profile to map for GraphQL
func mapDomainProfileToGraphQL(profile *domain.Profile, userID string) map[string]interface{} {
    if profile == nil {
        return nil
    }
    return map[string]interface{}{
        "user_id":   userID,
        "fullName":  profile.FullName,
        "email":     profile.Email,
        "phone":     profile.Phone,
        "birthDate": profile.BirthDate,
        "cep":       profile.CEP,
        "city":      profile.City,
        "state":     profile.State,
    }
}


// resolveWithTelemetry is a helper to wrap resolver execution with tracing and metrics
func (r *Resolver) resolveWithTelemetry(
	p graphql.ResolveParams,
	resolverName string,
	actualResolve func(p graphql.ResolveParams, span trace.Span) (interface{}, error),
) (interface{}, error) {
	startTime := time.Now()

    parentCtx := p.Context
    if parentCtx == nil {
        parentCtx = context.Background()
    }

	tracer := otel.Tracer(gqlTracerName)
	// Construct a more detailed span name if desired, e.g., "GraphQL Resolve: " + resolverName
	ctx, span := tracer.Start(parentCtx, resolverName, trace.WithAttributes(
		attribute.String("graphql.field.name", p.Info.FieldName),
        attribute.String("graphql.operation.type", string(p.Info.Operation.GetOperation())), // Cast OperationType to string
        attribute.String("graphql.path", fmt.Sprintf("%v",p.Info.Path.AsArray())),
	))
    p.Context = ctx

	defer span.End()

	var err error // Declare err here so it's in scope for the defer
	var result interface{}

	defer func() { // This defer will capture 'err' from the 'actualResolve' call
		statusCode := "OK"
		if err != nil {
			statusCode = "ERROR"
            span.RecordError(err)
            span.SetStatus(otelCodes.Error, err.Error())
		} else {
            span.SetStatus(otelCodes.Ok, "success")
        }

		commonAttrs := []attribute.KeyValue{
			attribute.String("graphql.resolver.name", resolverName), // Custom attribute
			attribute.String("graphql.field.name", p.Info.FieldName), // Standard or custom
			attribute.String("graphql.status_code", statusCode), // Custom status
		}
		r.GqlRequestsTotalCounter.Add(ctx, 1, metric.WithAttributes(commonAttrs...))
		duration := time.Since(startTime).Seconds()
		r.GqlRequestDurationSeconds.Record(ctx, duration, metric.WithAttributes(commonAttrs...))
	}()

	span.AddEvent(fmt.Sprintf("Executing %s resolver", resolverName))
	result, err = actualResolve(p, span)
	return result, err
}


// GetProfileResolver resolves the getProfile query.
func (r *Resolver) GetProfileResolver(p graphql.ResolveParams) (interface{}, error) {
	return r.resolveWithTelemetry(p, "GetProfileResolver", func(p graphql.ResolveParams, span trace.Span) (interface{}, error) {
		requestUserID, ok := p.Args["userId"].(string)
		if !ok || requestUserID == "" {
			return nil, errors.New("userId argument is required and must be a string")
		}
		span.SetAttributes(attribute.String("profile.user_id_requested", requestUserID))

		authUserID, ok := auth.UserIDFromContext(p.Context)
		if !ok || authUserID == "" {
			return nil, errors.New("unauthenticated: user ID not found in context")
		}
		span.SetAttributes(attribute.String("enduser.id_from_context", authUserID))


		if authUserID != requestUserID {
			return nil, errors.New("forbidden: you are not authorized to access this profile")
		}
        span.AddEvent("User authorized")

		profile, err := database.GetProfile(p.Context, r.FirestoreClient, requestUserID)
		if err != nil {
			dbStatus, _ := status.FromError(err)
			if dbStatus.Code() == gRPCCodes.NotFound { // Use aliased gRPCCodes
				return nil, fmt.Errorf("profile not found for user_id: %s", requestUserID)
			}
			return nil, fmt.Errorf("failed to get profile: %w", err)
		}
        span.AddEvent("Profile retrieved from database")
		return mapDomainProfileToGraphQL(profile, requestUserID), nil // Use corrected helper
	})
}
// CreateProfileResolver resolves the createProfile mutation.
func (r *Resolver) CreateProfileResolver(p graphql.ResolveParams) (interface{}, error) {
	return r.resolveWithTelemetry(p, "CreateProfileResolver", func(p graphql.ResolveParams, span trace.Span) (interface{}, error) {
		authUserID, ok := auth.UserIDFromContext(p.Context)
		if !ok || authUserID == "" {
			return nil, errors.New("unauthenticated: user ID not found in context")
		}
        span.SetAttributes(attribute.String("enduser.id_from_context", authUserID))

		inputMap, ok := p.Args["input"].(map[string]interface{})
		if !ok {
			return nil, errors.New("input argument is required and must be an object")
		}

		profileDomain := domain.Profile{}
        // Safely get required fields
        fn, fnOK := inputMap["fullName"].(string)
        em, emOK := inputMap["email"].(string)
        if !fnOK || fn == "" { return nil, errors.New("input.fullName is required") }
        if !emOK || em == "" { return nil, errors.New("input.email is required") }
        profileDomain.FullName = fn
        profileDomain.Email = em

        // Safely get optional fields
        if val, valOK := inputMap["phone"].(string); valOK { profileDomain.Phone = val }
        if val, valOK := inputMap["birthDate"].(string); valOK { profileDomain.BirthDate = val }
        if val, valOK := inputMap["cep"].(string); valOK { profileDomain.CEP = val }
        if val, valOK := inputMap["city"].(string); valOK { profileDomain.City = val }
        if val, valOK := inputMap["state"].(string); valOK { profileDomain.State = val }
        span.AddEvent("Input arguments processed")

		// TODO: Add validation for profileDomain using tags or a validation library

		err := database.CreateProfile(p.Context, r.FirestoreClient, authUserID, &profileDomain)
		if err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}
        span.AddEvent("Profile created in database")
		return mapDomainProfileToGraphQL(&profileDomain, authUserID), nil // Use corrected helper
	})
}

// UpdateProfileResolver resolves the updateProfile mutation.
func (r *Resolver) UpdateProfileResolver(p graphql.ResolveParams) (interface{}, error) {
    return r.resolveWithTelemetry(p, "UpdateProfileResolver", func(p graphql.ResolveParams, span trace.Span) (interface{}, error) {
        requestUserID, ok := p.Args["userId"].(string)
        if !ok || requestUserID == "" {
            return nil, errors.New("userId argument is required")
        }
        span.SetAttributes(attribute.String("profile.user_id_requested", requestUserID))

        authUserID, ok := auth.UserIDFromContext(p.Context)
        if !ok || authUserID == "" {
            return nil, errors.New("unauthenticated")
        }
        span.SetAttributes(attribute.String("enduser.id_from_context", authUserID))

        if authUserID != requestUserID {
            return nil, errors.New("forbidden")
        }
        span.AddEvent("User authorized")

        inputMap, ok := p.Args["input"].(map[string]interface{})
        if !ok {
            return nil, errors.New("input argument is required")
        }

        updateData := make(map[string]interface{})
        // Iterate over known profile fields to build updateData selectively
        // This avoids trying to update non-existent fields in Firestore if inputMap contains extra keys
        knownFields := []string{"fullName", "email", "phone", "birthDate", "cep", "city", "state"}
        for _, fieldName := range knownFields {
            if val, valOK := inputMap[fieldName]; valOK { // Check if field is present in input
                 updateData[fieldName] = val // Add it to map, Firestore handles type
            }
        }

        if len(updateData) == 0 {
            return nil, errors.New("no update fields provided in input")
        }
        span.AddEvent("Update map created", trace.WithAttributes(attribute.Int("graphql.input.update_field_count", len(updateData))))


        err := database.UpdateProfile(p.Context, r.FirestoreClient, requestUserID, updateData)
        if err != nil {
            return nil, fmt.Errorf("failed to update profile: %w", err)
        }
        span.AddEvent("Profile updated in database")

        updatedProfile, err := database.GetProfile(p.Context, r.FirestoreClient, requestUserID)
        if err != nil {
            return nil, fmt.Errorf("profile updated, but failed to retrieve: %w", err)
        }
        span.AddEvent("Updated profile retrieved")
        return mapDomainProfileToGraphQL(updatedProfile, requestUserID), nil // Use corrected helper
    })
}

// DeleteProfileResolver resolves the deleteProfile mutation.
func (r *Resolver) DeleteProfileResolver(p graphql.ResolveParams) (interface{}, error) {
    return r.resolveWithTelemetry(p, "DeleteProfileResolver", func(p graphql.ResolveParams, span trace.Span) (interface{}, error) {
        requestUserID, ok := p.Args["userId"].(string)
        if !ok || requestUserID == "" {
            return nil, errors.New("userId argument is required")
        }
        span.SetAttributes(attribute.String("profile.user_id_requested", requestUserID))

        authUserID, ok := auth.UserIDFromContext(p.Context)
        if !ok || authUserID == "" {
            return nil, errors.New("unauthenticated")
        }
        span.SetAttributes(attribute.String("enduser.id_from_context", authUserID))

        if authUserID != requestUserID {
            return nil, errors.New("forbidden")
        }
        span.AddEvent("User authorized")

        err := database.DeleteProfile(p.Context, r.FirestoreClient, requestUserID)
        if err != nil {
            return nil, fmt.Errorf("failed to delete profile: %w", err)
        }
        span.AddEvent("Profile deleted from database")
        // For delete, typically return the ID of the deleted object or a success boolean/message
        return map[string]interface{}{"user_id": requestUserID, "success": true}, nil
    })
}
```
