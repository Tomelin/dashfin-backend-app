package graphql

import (
	// "example.com/profile-service/internal/domain" // Not directly used here, but mapProfileToGraphQL implies its structure
	"github.com/graphql-go/graphql"
)

var profileType *graphql.Object
var createProfileInputType *graphql.InputObject
var updateProfileInputType *graphql.InputObject

// meterName for GraphQL specific metrics
const gqlHandlerMeterName = "example.com/profile-service/graphql-handler"

func InitSchemaTypes() {
	// Profile Type
	profileType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Profile",
		Fields: graphql.Fields{
			"user_id":   &graphql.Field{Type: graphql.NewNonNull(graphql.ID)}, // UserID from path/auth
			"fullName":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"email":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"phone":     &graphql.Field{Type: graphql.String},
			"birthDate": &graphql.Field{Type: graphql.String}, // Consider graphql.DateTime if specific format needed
			"cep":       &graphql.Field{Type: graphql.String},
			"city":      &graphql.Field{Type: graphql.String},
			"state":     &graphql.Field{Type: graphql.String},
		},
	})

	// Input type for creating a profile (UserID comes from context)
	createProfileInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "CreateProfileInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"fullName":  &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"email":     &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"phone":     &graphql.InputObjectFieldConfig{Type: graphql.String},
			"birthDate": &graphql.InputObjectFieldConfig{Type: graphql.String},
			"cep":       &graphql.InputObjectFieldConfig{Type: graphql.String},
			"city":      &graphql.InputObjectFieldConfig{Type: graphql.String},
			"state":     &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})

	// Input type for updating a profile
	updateProfileInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "UpdateProfileInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"fullName":  &graphql.InputObjectFieldConfig{Type: graphql.String},
			"email":     &graphql.InputObjectFieldConfig{Type: graphql.String}, // Usually email update is tricky
			"phone":     &graphql.InputObjectFieldConfig{Type: graphql.String},
			"birthDate": &graphql.InputObjectFieldConfig{Type: graphql.String},
			"cep":       &graphql.InputObjectFieldConfig{Type: graphql.String},
			"city":      &graphql.InputObjectFieldConfig{Type: graphql.String},
			"state":     &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})
}

// mapProfileToGraphQL helper is now defined in resolvers.go with correct domain.Profile typing
// and is named mapDomainProfileToGraphQL.
// This avoids an import of "example.com/profile-service/internal/domain" in this schema definition file,
// which is cleaner as schema definition itself doesn't need to know about the domain object, only its shape.
// The resolvers are responsible for the mapping.
