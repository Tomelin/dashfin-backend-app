package support

// SupportRequest defines the structure for a support request.
type SupportRequest struct {
	Category    string `json:"category" binding:"required"` // Enum: see constants.go
	Description string `json:"description" binding:"required,min=10,max=2000"`
}
