package http // Assuming this is internal/handler/http

import (
	"net/http" // Standard HTTP status codes
	"reflect" // For UpdateProfile to build map
	"strings" // For UpdateProfile to parse json tags

	"cloud.google.com/go/firestore"
	"example.com/profile-service/internal/database"
	"example.com/profile-service/internal/domain"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes" // For checking error codes from database layer
	"google.golang.org/grpc/status"
)

// ProfileHandler holds dependencies for profile HTTP handlers.
type ProfileHandler struct {
	FirestoreClient *firestore.Client
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(client *firestore.Client) *ProfileHandler {
	return &ProfileHandler{FirestoreClient: client}
}

// CreateProfile handles POST /profiles
// Note: The issue implies the userID for whom the profile is created is the authenticated user.
// The path /profiles suggests creating a profile for the authenticated user.
func (h *ProfileHandler) CreateProfile(c *gin.Context) {
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context, authentication required"})
		return
	}
	userIDStr, ok := authUserID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID in context is not a string"})
		return
	}

	var profile domain.Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Here you might add validation logic for the profile struct using validator tags if not done by ShouldBindJSON automatically
	// For example: err := validate.Struct(profile)

	err := database.CreateProfile(c.Request.Context(), h.FirestoreClient, userIDStr, &profile)
	if err != nil {
		// More specific error handling can be added based on possible errors from CreateProfile
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, profile) // Return the created profile
}

// GetProfile handles GET /profiles/:userId
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	requestUserID := c.Param("userId")
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context, authentication required"})
		return
	}

	if authUserID.(string) != requestUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this profile"})
		return
	}

	profile, err := database.GetProfile(c.Request.Context(), h.FirestoreClient, requestUserID)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// UpdateProfile handles PUT /profiles/:userId
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	requestUserID := c.Param("userId")
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context, authentication required"})
		return
	}
	if authUserID.(string) != requestUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this profile"})
		return
	}

	var updatesInput domain.Profile // Bind to the full profile struct to easily get all potential fields
	if err := c.ShouldBindJSON(&updatesInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

    // Create a map for non-zero value fields to update.
    // This is a common way to handle partial updates with structs.
    // More sophisticated methods might involve using pointers for all fields in a dedicated update struct,
    // or checking `c.Request.ParseForm()` and then `c.Request.PostForm` to see which fields were actually sent.
    // For simplicity with ShouldBindJSON and a full struct, we reflect.
	updateMap := make(map[string]interface{})
    val := reflect.ValueOf(updatesInput)
    typ := val.Type()
    for i := 0; i < val.NumField(); i++ {
        field := typ.Field(i)
        value := val.Field(i)
        // Check if the field is non-zero. This means if a user wants to set an optional field to its zero value (e.g. empty string),
        // this approach won't work directly. For such cases, a map[string]interface{} in request is better,
        // or pointer fields in the struct. Given the YAML, optional fields are common.
        // For now, we proceed with this, but it's a known limitation for explicitly setting to zero.
        // We only add fields that are non-zero in the input.
        // An alternative would be to bind to map[string]interface{} directly if all fields are optional in update.
        if !reflect.DeepEqual(value.Interface(), reflect.Zero(field.Type).Interface()) {
             jsonTag := field.Tag.Get("json")
             // Handle cases like "fullName,omitempty"
             tagName := strings.Split(jsonTag, ",")[0]
             if tagName != "" && tagName != "-" {
                updateMap[tagName] = value.Interface()
             } else if field.Name != "" { // Fallback to field name if no json tag or it's weird
                updateMap[typ.Field(i).Name] = value.Interface()
             }
        }
    }


	if len(updateMap) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No update fields provided or all fields are zero values"})
		return
	}


	err := database.UpdateProfile(c.Request.Context(), h.FirestoreClient, requestUserID, updateMap)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found for update"})
			return
		}
		if st.Code() == codes.InvalidArgument {
             c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
             return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

    // Fetch the updated profile to return it
    updatedProfile, getErr := database.GetProfile(c.Request.Context(), h.FirestoreClient, requestUserID)
    if getErr != nil {
        c.JSON(http.StatusOK, gin.H{"message": "Profile updated, but failed to retrieve updated version."})
        return
    }
	c.JSON(http.StatusOK, updatedProfile)
}

// DeleteProfile handles DELETE /profiles/:userId
func (h *ProfileHandler) DeleteProfile(c *gin.Context) {
	requestUserID := c.Param("userId")
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context, authentication required"})
		return
	}

	if authUserID.(string) != requestUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this profile"})
		return
	}

	err := database.DeleteProfile(c.Request.Context(), h.FirestoreClient, requestUserID)
	if err != nil {
		// Delete in Firestore doesn't typically error if not found, but other errors can occur.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile: " + err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
