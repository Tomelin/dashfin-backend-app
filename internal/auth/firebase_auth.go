package auth

import (
	"context"
	"log"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth" // Ensure this is imported for token.UID
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"example.com/profile-service/internal/config" // For AppConfig

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "example.com/profile-service/auth" // Define a tracer name

// Define a custom context key for userID
type contextKey string

const userIDKey contextKey = "userID"

// NewContextWithUserID creates a new context with the userID value.
func NewContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext retrieves the userID from context.
// Returns empty string if not found.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// InitFirebase initializes the Firebase Admin SDK.
// It expects the GOOGLE_APPLICATION_CREDENTIALS environment variable to be set.
func InitFirebase(ctx context.Context) (*firebase.App, error) {
	conf := &firebase.Config{
		// ProjectID will be inferred from the credentials if not set.
	}
	// If GOOGLE_APPLICATION_CREDENTIALS is set, it's used automatically.
	// Otherwise, specify credentials using option.WithCredentialsFile("path/to/serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		// log.Fatalf causes exit. For a library function, better to return the error.
		log.Printf("error initializing app: %v\n", err)
		return nil, err
	}
	return app, nil
}

// FirebaseAuthMiddleware creates a Gin middleware for Firebase authentication and OTel tracing.
func FirebaseAuthMiddleware(firebaseApp *firebase.App, appCfg *config.Config) gin.HandlerFunc { // Pass AppConfig
	return func(c *gin.Context) {
		// OTel: Extract trace context from headers
		propagator := otel.GetTextMapPropagator()
		otelCtx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// OTel: Try to get X-TRACE-ID if no traceparent was found
		if !trace.SpanFromContext(otelCtx).SpanContext().HasTraceID() && appCfg.OpenTelemetry.TraceHeaderName != "" {
			xtraceID := c.GetHeader(appCfg.OpenTelemetry.TraceHeaderName)
			if xtraceID != "" {
				// Attempt to parse X-TRACE-ID as a W3C TraceID.
				parsedTraceID, err := trace.TraceIDFromHex(strings.ReplaceAll(xtraceID, "-", "")) // Common format for trace IDs
				if err == nil {
					spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
						TraceID:    parsedTraceID,
						SpanID:     trace.SpanID{}, // Will be generated
						TraceFlags: trace.FlagsSampled, // Or based on upstream decision if available
					})
					otelCtx = trace.ContextWithRemoteSpanContext(otelCtx, spanCtx)
				} else {
					log.Printf("Could not parse %s header value '%s' as TraceID: %v. Starting new trace.", appCfg.OpenTelemetry.TraceHeaderName, xtraceID, err)
				}
			}
		}

		tracer := otel.Tracer(tracerName)
		spanName := "HTTP " + c.Request.Method + " " + c.FullPath() // Use c.FullPath() for parameterized route
		if c.FullPath() == "" { // Handle cases where FullPath might not be available (e.g. 404)
			spanName = "HTTP " + c.Request.Method + " " + c.Request.URL.Path
		}
		var span trace.Span
		otelCtx, span = tracer.Start(otelCtx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(c.Request.Method),
				semconv.HTTPURLKey.String(c.Request.URL.String()),
				semconv.HTTPTargetKey.String(c.Request.URL.Path),
				semconv.HTTPRouteKey.String(c.FullPath()), // Gin specific route
				semconv.UserAgentOriginalKey.String(c.Request.UserAgent()),
				semconv.NetHostNameKey.String(c.Request.URL.Host), // Using NetHostNameKey as per recent semconv for server hostname
			),
		)
		// HTTPSchemeKey is not directly available in c.Request.URL.Scheme for server side if not behind a reverse proxy properly setting it.
		// It can be "http" or "https"
		if c.Request.TLS != nil {
			span.SetAttributes(semconv.HTTPSchemeKey.String("https"))
		} else {
			span.SetAttributes(semconv.HTTPSchemeKey.String("http"))
		}
		defer span.End()

		// Store OTel context in Gin's request context
		c.Request = c.Request.WithContext(otelCtx)

		// --- Original Auth Logic ---
		authToken := c.GetHeader("X-AUTHORIZATION")
		appHeader := c.GetHeader("X-APP")
		userIDHeader := c.GetHeader("X-USERID")

		if authToken == "" {
			span.SetStatus(codes.Unauthenticated, "X-AUTHORIZATION header required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-AUTHORIZATION header required"})
			return
		}
		if appHeader == "" {
			span.SetStatus(codes.Unauthenticated, "X-APP header required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-APP header required"})
			return
		}
		if userIDHeader == "" {
			span.SetStatus(codes.Unauthenticated, "X-USERID header required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-USERID header required"})
			return
		}

		// Remove "Bearer " prefix if present
		if strings.HasPrefix(authToken, "Bearer ") {
			authToken = strings.TrimPrefix(authToken, "Bearer ")
		}

		client, err := firebaseApp.Auth(otelCtx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Internal, "Error getting Auth client")
			log.Printf("Error getting Auth client: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error initializing auth"})
			return
		}
		token, err := client.VerifyIDToken(otelCtx, authToken) // Pass otelCtx
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Unauthenticated, "Invalid X-AUTHORIZATION token")
			log.Printf("Error verifying ID token: %v\n", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid X-AUTHORIZATION token"})
			return
		}
		if token.UID != userIDHeader {
			span.SetStatus(codes.Unauthenticated, "X-USERID does not match token UID")
			log.Printf("X-USERID (%s) does not match token UID (%s)\n", userIDHeader, token.UID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-USERID does not match authenticated user"})
			return
		}
		// --- End of Original Auth Logic ---

		// Add UserID to span attributes after successful auth
		span.SetAttributes(attribute.String("enduser.id", token.UID))
		span.SetAttributes(attribute.String("app.name", appHeader))


		c.Set("userID", token.UID) // Store validated userID in Gin context
		c.Set("otelContext", otelCtx) // Store otel context for handlers to use

		c.Next()

		// After request is handled, set HTTP status code on the span
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(c.Writer.Status()))
		// Set span status based on HTTP status code
        if c.Writer.Status() >= http.StatusInternalServerError {
            span.SetStatus(codes.Internal, "Server error")
        } else if c.Writer.Status() >= http.StatusBadRequest {
            span.SetStatus(codes.InvalidArgument, "Client error")
        } else {
            span.SetStatus(codes.OK, "")
        }
	}
}

// MetadataCarrier adapts gRPC metadata.MD to TextMapCarrier for OTel propagation.
type MetadataCarrier struct {
	MD metadata.MD
}

// Get returns the value associated with the passed key.
func (mc MetadataCarrier) Get(key string) string {
	vals := mc.MD.Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// Set stores the key-value pair.
func (mc MetadataCarrier) Set(key string, value string) {
	// Ensure MD is not nil, especially if it was an empty MD from FromIncomingContext
	if mc.MD == nil {
		mc.MD = metadata.MD{}
	}
	mc.MD.Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (mc MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc.MD))
	for k := range mc.MD {
		keys = append(keys, k)
	}
	return keys
}

// UnaryAuthInterceptor provides gRPC unary server interceptor for Firebase authentication and OTel tracing.
func UnaryAuthInterceptor(firebaseApp *firebase.App, appCfg *config.Config) grpc.UnaryServerInterceptor { // Pass AppConfig
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		// OTel: Extract trace context from metadata
		propagator := otel.GetTextMapPropagator()
		otelCtx := propagator.Extract(ctx, MetadataCarrier{MD: md})

		// OTel: Try to get X-TRACE-ID if no traceparent was found
        if !trace.SpanFromContext(otelCtx).SpanContext().HasTraceID() && appCfg.OpenTelemetry.TraceHeaderName != "" {
			// gRPC metadata keys are typically lowercase.
            xtraceIDs := md.Get(strings.ToLower(appCfg.OpenTelemetry.TraceHeaderName))
            if len(xtraceIDs) > 0 {
                xtraceID := xtraceIDs[0]
                parsedTraceID, err := trace.TraceIDFromHex(strings.ReplaceAll(xtraceID, "-", ""))
                if err == nil {
					spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
						TraceID:    parsedTraceID,
						SpanID:     trace.SpanID{},
						TraceFlags: trace.FlagsSampled,
					})
					otelCtx = trace.ContextWithRemoteSpanContext(otelCtx, spanCtx)
                } else {
                    log.Printf("gRPC: Could not parse %s header value '%s' as TraceID: %v. Starting new trace.", appCfg.OpenTelemetry.TraceHeaderName, xtraceID, err)
                }
            }
        }

		tracer := otel.Tracer(tracerName)
		spanName := info.FullMethod // gRPC method name
		var span trace.Span
		otelCtx, span = tracer.Start(otelCtx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.RPCSystemKey.String("grpc"),
				semconv.RPCServiceKey.String(info.FullMethod),
			),
		)
		defer span.End()

		// --- Original Auth Logic ---
		authHeaders := md.Get("x-authorization")
		appHeaders := md.Get("x-app")
		userIDHeaders := md.Get("x-userid")

		var appHeader string
		if len(appHeaders) > 0 {
			appHeader = appHeaders[0]
		} else {
			span.SetStatus(codes.Unauthenticated, "x-app header required")
			return nil, status.Errorf(codes.Unauthenticated, "x-app header required")
		}

		var userIDHeader string
		if len(userIDHeaders) > 0 {
			userIDHeader = userIDHeaders[0]
		} else {
			span.SetStatus(codes.Unauthenticated, "x-userid header required")
			return nil, status.Errorf(codes.Unauthenticated, "x-userid header required")
		}

		var authToken string
		if len(authHeaders) > 0 {
			authToken = authHeaders[0]
			if strings.HasPrefix(authToken, "Bearer ") {
				authToken = strings.TrimPrefix(authToken, "Bearer ")
			}
		} else {
			span.SetStatus(codes.Unauthenticated, "x-authorization header required")
			return nil, status.Errorf(codes.Unauthenticated, "x-authorization header required")
		}


		client, errClient := firebaseApp.Auth(otelCtx)
		if errClient != nil {
			span.RecordError(errClient)
			span.SetStatus(codes.Internal, "Error getting Auth client for gRPC")
			log.Printf("Error getting Auth client for gRPC: %v\n", errClient)
			return nil, status.Errorf(codes.Internal, "Internal server error initializing auth")
		}

		token, errToken := client.VerifyIDToken(otelCtx, authToken)
		if errToken != nil {
			span.RecordError(errToken)
			span.SetStatus(codes.Unauthenticated, "Invalid x-authorization token")
			log.Printf("Error verifying ID token for gRPC: %v\n", errToken)
			return nil, status.Errorf(codes.Unauthenticated, "Invalid x-authorization token: %v", errToken)
		}

		if token.UID != userIDHeader {
			span.SetStatus(codes.Unauthenticated, "x-userid does not match token UID")
			log.Printf("gRPC: X-USERID (%s) does not match token UID (%s)\n", userIDHeader, token.UID)
			return nil, status.Errorf(codes.Unauthenticated, "x-userid does not match authenticated user")
		}
		// --- End of Original Auth Logic ---

		span.SetAttributes(attribute.String("enduser.id", token.UID))
		span.SetAttributes(attribute.String("app.name", appHeader))

		finalCtx := NewContextWithUserID(otelCtx, token.UID)
		resp, err := handler(finalCtx, req)

		// Set span status based on gRPC error
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(s.Code(), s.Message())
			span.RecordError(err)
		} else {
			span.SetStatus(codes.OK, "")
		}
		span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(status.Code(err))))

		return resp, err
	}
}
