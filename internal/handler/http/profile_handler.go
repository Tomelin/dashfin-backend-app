package http // Assuming this is internal/handler/http

import (
	"net/http" // Standard HTTP status codes
	"reflect"  // For UpdateProfile to build map
	"strings"  // For UpdateProfile to parse json tags
	"time"     // For timing requests

	"cloud.google.com/go/firestore"
	"example.com/profile-service/internal/database"
	"example.com/profile-service/internal/domain"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes" // For checking error codes from database layer
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const httpHandlerMeterName = "example.com/profile-service/http-handler"

// ProfileHandler holds dependencies for profile HTTP handlers.
type ProfileHandler struct {
	FirestoreClient            *firestore.Client
	HttpRequestsTotalCounter   metric.Int64Counter
	HttpRequestDurationSeconds metric.Float64Histogram
	// Add more metrics as needed
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(client *firestore.Client) *ProfileHandler {
	meter := otel.Meter(httpHandlerMeterName)
	requestsCounter, errCounter := meter.Int64Counter("http.server.requests_total",
		metric.WithDescription("Total number of HTTP requests."),
		metric.WithUnit("{request}"),
	)
	durationHistogram, errHist := meter.Float64Histogram("http.server.duration_seconds",
		metric.WithDescription("HTTP request duration in seconds."),
		metric.WithUnit("s"),
	)
	if errCounter != nil {
		otel.Handle(errCounter) // Use global error handler
	}
	if errHist != nil {
		otel.Handle(errHist)
	}

	return &ProfileHandler{
		FirestoreClient:            client,
		HttpRequestsTotalCounter:   requestsCounter,
		HttpRequestDurationSeconds: durationHistogram,
	}
}

// CreateProfile handles POST /profiles
// Note: The issue implies the userID for whom the profile is created is the authenticated user.
// The path /profiles suggests creating a profile for the authenticated user.
func (h *ProfileHandler) CreateProfile(c *gin.Context) {
	startTime := time.Now()
	// otelReqCtx is the context from the request, potentially updated by auth middleware with OTel span
	otelReqCtx := c.Request.Context()
	span := trace.SpanFromContext(otelReqCtx)
	span.AddEvent("Handling CreateProfile request")

	defer func() {
		// It's important that c.Writer.Status() is called after the handler logic has mostly run.
		// In Gin, status is typically set when c.JSON, c.Status, etc. are called.
		// If the handler panics or doesn't explicitly write a status, it might be 0 or 200 by default.
		// The auth middleware already sets span status based on final response, this adds metric attributes.
		statusCode := c.Writer.Status()
		if statusCode == 0 && len(c.Errors) > 0 { // If status is 0 but there were errors, likely 500
			statusCode = http.StatusInternalServerError
		} else if statusCode == 0 { // If no errors and status 0, assume 200 if not set
			statusCode = http.StatusOK
		}

		commonAttrs := []attribute.KeyValue{
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.method", c.Request.Method),
			attribute.Int("http.status_code", statusCode),
		}
		h.HttpRequestsTotalCounter.Add(otelReqCtx, 1, metric.WithAttributes(commonAttrs...))
		duration := time.Since(startTime).Seconds()
		h.HttpRequestDurationSeconds.Record(otelReqCtx, duration, metric.WithAttributes(commonAttrs...))
	}()

	authUserID, exists := c.Get("userID")
	if !exists {
		// Status will be set by c.JSON, defer will pick it up
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context, authentication required"})
		return
	}
	userIDStr, ok := authUserID.(string)
	if !ok {
		// Status will be set by c.JSON, defer will pick it up
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID in context is not a string"})
		return
	}

	var profile domain.Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		// Status will be set by c.JSON, defer will pick it up
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Here you might add validation logic for the profile struct using validator tags if not done by ShouldBindJSON automatically
	// For example: err := validate.Struct(profile)
	span.AddEvent("Request payload bound and validated")

	err := database.CreateProfile(otelReqCtx, h.FirestoreClient, userIDStr, &profile)
	if err != nil {
		span.RecordError(err)
		// More specific error handling can be added based on possible errors from CreateProfile
		// Status will be set by c.JSON, defer will pick it up
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile: " + err.Error()})
		return
	}
	span.AddEvent("Profile created in database")
	c.JSON(http.StatusCreated, profile) // Return the created profile
}

// GetProfile handles GET /profiles/:userId
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	startTime := time.Now()
	otelReqCtx := c.Request.Context()
	span := trace.SpanFromContext(otelReqCtx)
	span.AddEvent("Handling GetProfile request")

	defer func() {
		statusCode := c.Writer.Status()
		if statusCode == 0 && len(c.Errors) > 0 {
			statusCode = http.StatusInternalServerError
		} else if statusCode == 0 {
			statusCode = http.StatusOK
		}
		commonAttrs := []attribute.KeyValue{
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.method", c.Request.Method),
			attribute.Int("http.status_code", statusCode),
		}
		h.HttpRequestsTotalCounter.Add(otelReqCtx, 1, metric.WithAttributes(commonAttrs...))
		duration := time.Since(startTime).Seconds()
		h.HttpRequestDurationSeconds.Record(otelReqCtx, duration, metric.WithAttributes(commonAttrs...))
	}()

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
	span.SetAttributes(attribute.String("profile.user_id", requestUserID))

	profile, err := database.GetProfile(otelReqCtx, h.FirestoreClient, requestUserID)
	if err != nil {
		span.RecordError(err)
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile: " + err.Error()})
		return
	}
	span.AddEvent("Profile retrieved from database")
	c.JSON(http.StatusOK, profile)
}

// UpdateProfile handles PUT /profiles/:userId
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	startTime := time.Now()
	otelReqCtx := c.Request.Context()
	span := trace.SpanFromContext(otelReqCtx)
	span.AddEvent("Handling UpdateProfile request")

	defer func() {
		statusCode := c.Writer.Status()
		if statusCode == 0 && len(c.Errors) > 0 {
			statusCode = http.StatusInternalServerError
		} else if statusCode == 0 {
			statusCode = http.StatusOK
		}
		commonAttrs := []attribute.KeyValue{
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.method", c.Request.Method),
			attribute.Int("http.status_code", statusCode),
		}
		h.HttpRequestsTotalCounter.Add(otelReqCtx, 1, metric.WithAttributes(commonAttrs...))
		duration := time.Since(startTime).Seconds()
		h.HttpRequestDurationSeconds.Record(otelReqCtx, duration, metric.WithAttributes(commonAttrs...))
	}()

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
	span.SetAttributes(attribute.String("profile.user_id", requestUserID))

	var updatesInput domain.Profile // Bind to the full profile struct to easily get all potential fields
	if err := c.ShouldBindJSON(&updatesInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}
	span.AddEvent("Request payload bound")

	updateMap := make(map[string]interface{})
	val := reflect.ValueOf(updatesInput)
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		if !reflect.DeepEqual(value.Interface(), reflect.Zero(field.Type).Interface()) {
			jsonTag := field.Tag.Get("json")
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName != "" && tagName != "-" {
				updateMap[tagName] = value.Interface()
			} else if field.Name != "" {
				updateMap[typ.Field(i).Name] = value.Interface()
			}
		}
	}
	span.AddEvent("Update map created", trace.WithAttributes(attribute.Int("update_map.size", len(updateMap))))

	if len(updateMap) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No update fields provided or all fields are zero values"})
		return
	}

	err := database.UpdateProfile(otelReqCtx, h.FirestoreClient, requestUserID, updateMap)
	if err != nil {
		span.RecordError(err)
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
	span.AddEvent("Profile updated in database")

	updatedProfile, getErr := database.GetProfile(otelReqCtx, h.FirestoreClient, requestUserID)
	if getErr != nil {
		span.RecordError(getErr)
		// Status already set by previous successful update, but we couldn't retrieve the latest.
		// This is an inconsistency state, but the update itself was successful.
		// Client will get the "updated but failed to retrieve" message.
		c.JSON(http.StatusOK, gin.H{"message": "Profile updated, but failed to retrieve updated version."})
		return
	}
	span.AddEvent("Updated profile retrieved")
	c.JSON(http.StatusOK, updatedProfile)
}

// DeleteProfile handles DELETE /profiles/:userId
func (h *ProfileHandler) DeleteProfile(c *gin.Context) {
	startTime := time.Now()
	otelReqCtx := c.Request.Context()
	span := trace.SpanFromContext(otelReqCtx)
	span.AddEvent("Handling DeleteProfile request")

	defer func() {
		statusCode := c.Writer.Status()
		if statusCode == 0 && len(c.Errors) > 0 {
			statusCode = http.StatusInternalServerError
		} else if statusCode == 0 { // No content is 204, but if not set, might be 200.
			statusCode = http.StatusNoContent // Default for successful delete if not set otherwise
		}
		commonAttrs := []attribute.KeyValue{
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.method", c.Request.Method),
			attribute.Int("http.status_code", statusCode),
		}
		h.HttpRequestsTotalCounter.Add(otelReqCtx, 1, metric.WithAttributes(commonAttrs...))
		duration := time.Since(startTime).Seconds()
		h.HttpRequestDurationSeconds.Record(otelReqCtx, duration, metric.WithAttributes(commonAttrs...))
	}()

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
	span.SetAttributes(attribute.String("profile.user_id", requestUserID))

	err := database.DeleteProfile(otelReqCtx, h.FirestoreClient, requestUserID)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile: " + err.Error()})
		return
	}
	span.AddEvent("Profile deleted from database")
	c.Status(http.StatusNoContent)
}
