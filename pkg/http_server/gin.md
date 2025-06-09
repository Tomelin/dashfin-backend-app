# Gin-Gonic HTTP Server Library Documentation

## 1. Introduction

This document provides guidance on using the existing Gin-Gonic based HTTP server library within this project. The library is designed to facilitate the creation of RESTful APIs and already includes common functionalities like configuration management, logging, telemetry, and CORS.

The primary files for the server setup are:
- `pkg/http_server/config.go`: Contains configuration structures (`RestAPIConfig`, `RestAPI`) and the `NewRestApi` function for server initialization.
- `pkg/http_server/server.go`: Contains server runtime logic, including middleware.

This guide will focus on how to add new functionalities, specifically CRUD (Create, Read, Update, Delete) operations for new resources, using "registros" (records) as a primary example.

## 2. Core Concepts

### 2.1. Server Initialization

The server is initialized using the `NewRestApi` function from `pkg/http_server/config.go`. This function takes a `RestAPIConfig` struct as input, which defines settings like port, host, version, and SSL.

```go
// Example of RestAPIConfig (defined in pkg/http_server/config.go)
type RestAPIConfig struct {
    Port            string `mapstructure:"port"`
    SSLEnabled      bool   `mapstructure:"ssl_enabled"`
    Host            string `mapstructure:"host"`
    Version         string `mapstructure:"version"`
    Name            string `mapstructure:"name"`
    // ... other fields
}

// Initialization (simplified example)
// import "your_project_path/pkg/http_server"
//
// func main() {
//     config := http_server.RestAPIConfig{
//         Port:    "8080",
//         Host:    "localhost",
//         Name:    "myapi",
//         Version: "v1",
//     }
//     api, err := http_server.NewRestApi(config)
//     if err != nil {
//         log.Fatalf("Failed to initialize API: %v", err)
//     }
//     // api.Route is the *gin.Engine
//     // api.RouterGroup is a *gin.RouterGroup (e.g., /myapi)
//     // ... add routes ...
//     api.Run(nil) // Or your main handler if not using Gin's default
// }
```

### 2.2. Default Middleware

The `newRestAPI` function in `config.go` already sets up several useful middleware by default:
- **OpenTelemetry (`otelgin.Middleware`)**: For distributed tracing.
- **Logger (`gin.Logger()`)**: For request logging.
- **Recovery (`gin.Recovery()`)**: To recover from panics and return a 500 error.
- **CORS**: A permissive CORS policy is set up by default (see section 5).
- **Swagger**: Endpoints for API documentation (`/docs/swagger/*any`) are also configured.

## 3. Adding New CRUD Endpoints (Example: "Registros")

To add CRUD operations for a new resource like "registros", you'll define routes and handler functions.

### 3.1. Defining Routes

Routes are typically defined on the `api.RouterGroup` or a sub-group created from it.

```go
// Assume 'api' is your *http_server.RestAPI instance
// api.RouterGroup is typically something like /<api.Config.Name> (e.g., /myapi)

// You can create further groups for versioning or resource organization
// For example, if api.RouterGroup points to /myapi, then v1 will be /myapi/v1
v1 := api.RouterGroup.Group(fmt.Sprintf("/%s", api.Config.Version)) // e.g., /v1 if api.Config.Version is "v1"

registrosGroup := v1.Group("/registros")
{
    registrosGroup.POST("/", CreateRegistroHandler)          // Create a new registro
    registrosGroup.GET("/", GetRegistrosHandler)             // Get all registros (with potential filters)
    registrosGroup.GET("/:id", GetRegistroByIDHandler)       // Get a specific registro by ID
    registrosGroup.PUT("/:id", UpdateRegistroHandler)        // Update a specific registro by ID
    registrosGroup.DELETE("/:id", DeleteRegistroHandler)     // Delete a specific registro by ID
}
```
**Note:** The `api.RouterGroup` (returned by `http_server.NewRestApi`) is already established by the `newRestAPI` internal function and corresponds to the path `/<config.Name>` (e.g., if `config.Name` is "myapi", the base path is `/myapi`). The subsequent line `v1 := api.RouterGroup.Group(fmt.Sprintf("/%s", api.Config.Version))` creates a new group under this base path. So, if `api.Config.Name` is "myapi" and `api.Config.Version` is "v1", the full path for `registrosGroup` (`v1.Group("/registros")`) would be `/myapi/v1/registros`.

### 3.2. Handler Functions

Handler functions in Gin receive a `*gin.Context` which provides access to request details and methods to write the response.

```go
// Example handler signatures (implementation details omitted)

// import "github.com/gin-gonic/gin"

func CreateRegistroHandler(c *gin.Context) {
    // Logic to bind request body, validate, create a new registro
    // c.JSON(http.StatusCreated, newRegistro)
}

func GetRegistrosHandler(c *gin.Context) {
    // Logic to get query parameters for filtering, fetch registros
    // c.JSON(http.StatusOK, listOfRegistros)
}

func GetRegistroByIDHandler(c *gin.Context) {
    // id := c.Param("id")
    // Logic to fetch registro by ID
    // c.JSON(http.StatusOK, registro)
}

func UpdateRegistroHandler(c *gin.Context) {
    // id := c.Param("id")
    // Logic to bind request body, validate, update registro
    // c.JSON(http.StatusOK, updatedRegistro)
}

func DeleteRegistroHandler(c *gin.Context) {
    // id := c.Param("id")
    // Logic to delete registro
    // c.Status(http.StatusNoContent)
}
```

### 3.3. Structuring Handlers

For better organization, it's recommended to place handler functions in separate packages. For example:
- `internal/registros/handlers.go`
- `pkg/registros/http_handlers.go` (if exposing them as part of a reusable package)

## 4. Routing and Middleware Application

### 4.1. `router.Group()`

As seen above, `router.Group()` is used to organize routes. This is useful for:
- **Versioning**: `v1 := api.RouterGroup.Group("/v1")`, `v2 := api.RouterGroup.Group("/v2")`
- **Resource Grouping**: `userGroup := v1.Group("/users")`, `productGroup := v1.Group("/products")`

Middleware can be applied to these groups.

### 4.2. `router.USE()`

The `router.Use()` function is used to apply middleware globally or to a specific group. Middleware added with `Use()` will be executed for all subsequent routes defined on that router or group.

```go
// Global middleware (applied to all routes on `api.Route` *gin.Engine)
// api.Route.Use(MyGlobalMiddleware())

// Group-specific middleware
// This middleware will only apply to routes within 'registrosGroup'
// and any sub-groups of 'registrosGroup'.
registrosGroup := api.RouterGroup.Group(fmt.Sprintf("/%s/registros", api.Config.Version))
registrosGroup.Use(RegistrosSpecificMiddleware())
{
    registrosGroup.POST("/", CreateRegistroHandler)
    // ... other registro routes
}
```
The existing `newRestAPI` function in `config.go` already uses `router.Use()` to apply global middleware like `otelgin.Middleware`, `gin.Logger()`, and `gin.Recovery()`.

## 5. Middlewares

Middleware functions in Gin are handlers that can execute code before or after the main request handler. They have access to the `*gin.Context` and can:
- Modify request/response objects.
- Perform authentication/authorization checks.
- Log requests/responses.
- Handle errors.
- Halt the request chain (e.g., if authentication fails).

### 5.1. Writing Custom Middleware

A middleware is essentially a `gin.HandlerFunc`.

```go
func MyCustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Code to execute before the request handler
        fmt.Println("Middleware: Before request")

        // Pass control to the next handler in the chain
        c.Next() // This is crucial!

        // Code to execute after the request handler
        fmt.Println("Middleware: After request")
    }
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("X-Auth-Token")
        if token == "" || !isValidToken(token) { // isValidToken is a placeholder
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return // Abort further processing
        }
        // Optionally, set user information in the context
        // c.Set("userID", "someUserID")
        c.Next()
    }
}
```

### 5.2. Applying Middleware

- **Globally or to a group:** `router.Use(MyCustomMiddleware())`
- **To a specific route:** `router.GET("/secure/data", AuthMiddleware(), GetDataHandler)`

The `server.go` file contains `MiddlewareHeader` which is an example of a custom middleware for checking specific headers:
```go
// From pkg/http_server/server.go
func (s *RestAPI) MiddlewareHeader(c *gin.Context) {
    if c.GetHeader("X-USERID") == "" {
        c.AbortWithStatus(401) // Consider gin.H for JSON error
        return
    }
    if c.GetHeader("X-AUTHORIZATION") == "" { // Typo in original? X-Authorization vs X-AUTHORIZATION
        c.AbortWithStatus(401) // Consider gin.H for JSON error
        return
    }
    c.Next()
}
```
This middleware can be applied using `api.RouterGroup.Use(api.MiddlewareHeader)` if you want to use the `RestAPI` instance `s` or by adapting it if `s` is not available in the scope where routes are defined. If `MiddlewareHeader` doesn't depend on `s`, it can be a standalone `gin.HandlerFunc`.

## 6. CORS (Cross-Origin Resource Sharing)

### 6.1. What is CORS?

CORS is a browser security feature that restricts cross-origin HTTP requests initiated from scripts. When your frontend (running on one domain) tries to make an API call to your backend (running on another domain), CORS policies on the server dictate whether the browser should allow this.

### 6.2. Default Configuration

The `newRestAPI` function in `pkg/http_server/config.go` sets up a very permissive CORS policy using `github.com/gin-contrib/cors`:

```go
// In config.go's newRestAPI function:
corsConfig := cors.DefaultConfig()
corsConfig.AllowAllOrigins = true // Allows requests from any origin
corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", " X-Authorization", " X-USERID", " X-APP", " X-USER", " X-TRACE-ID"}
router.Use(cors.New(corsConfig)) // Applied globally
```
This default configuration is suitable for many development scenarios and some production environments where the API is intended to be public.

### 6.3. Customizing CORS

If you need a more restrictive CORS policy (e.g., allowing only specific origins), you would modify the `cors.Config` structure in `pkg/http_server/config.go` before `cors.New(corsConfig)` is called.

Example of a more restrictive policy:
```go
// Hypothetical modification in config.go if needed
// corsConfig.AllowAllOrigins = false
// corsConfig.AllowOrigins = []string{"https://meudominio.com", "https://app.meudominio.com"}
// corsConfig.AllowMethods = []string{"GET", "POST", "PUT"}
// ... other specific settings
```
**Important:** Since the task is to *not alter existing Go files*, this documentation will assume the existing permissive CORS policy is in use. If a different policy is needed for specific route groups, Gin allows applying different CORS middleware instances to different groups, but this would typically involve instantiating `cors.New()` with a new config for that group.

The `server.go` file also contains a `CorsMiddleware` method on the `RestAPI` struct:
```go
// From pkg/http_server/server.go
func (s *RestAPI) CorsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"}, // Note: "*" is often less explicit than listing methods
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", " X-Authorization", " X-USERID", " X-APP", " X-USER", " X-TRACE-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
```
This method provides another, slightly different, permissive CORS configuration. If you wanted to apply this specific configuration (perhaps it's more up-to-date or has specific credential settings you need), you could apply it using `api.Route.Use(api.CorsMiddleware())` or to a specific group, assuming `api` is your `*RestAPI` instance. Given that `config.go` already sets a global CORS policy, applying another one globally might be redundant or lead to unexpected behavior unless the first one is removed or carefully managed. Typically, one global CORS policy is preferred.

**Recommendation:** Stick to the globally defined CORS policy in `config.go` for simplicity. If highly specific CORS rules are needed for a subset of routes that *differ* from the global policy, you could apply a new `cors.New(customConfig)` to that specific router group.

## 7. Recommendations & Best Practices

### 7.1. Error Handling
- Use `c.AbortWithStatusJSON(statusCode, gin.H{"error": "message"})` for sending structured error responses.
- Define standard error response formats.
- Consider a centralized error handling middleware.

### 7.2. Input Validation
- Validate request bodies and parameters early in your handlers.
- Use libraries like `go-playground/validator` for struct validation.
- Return clear error messages for invalid input (e.g., `400 Bad Request`).

### 7.3. Project Structure
- **Separation of Concerns**:
    - **Handlers (e.g., `internal/registros/handlers.go`)**: Responsible for parsing requests, calling services, and formatting responses. Keep them thin.
    - **Services (e.g., `internal/registros/service.go`)**: Contain business logic.
    - **Repositories (e.g., `internal/registros/repository.go`)**: Handle data persistence (database interactions).
- This structure promotes testability and maintainability.

### 7.4. Dependency Injection
- Pass dependencies (like database connections, services) to your handlers/services rather than using global variables. This improves testability and modularity.
- Consider using a dependency injection framework for larger applications if preferred.

### 7.5. Configuration
- Leverage the existing `RestAPIConfig` for server-related configurations.
- For other application configurations, consider using libraries like Viper, and load them at startup.

### 7.6. Testing
- Write unit tests for handlers, services, and repositories.
- Use Go's built-in `net/http/httptest` package for testing Gin handlers.

This documentation should provide a solid foundation for extending the API with new features like "registros" while adhering to the established patterns in the `pkg/http_server` library.
