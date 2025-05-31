package domain

// Profile defines the structure for user profile data.
type Profile struct {
    // ID an    type: stringd Firebase UID will be used as the document ID in Firestore.
    // It's not a field in the struct that's stored, but rather the key for the document.
    // UserID      string `json:"id,omitempty" firestore:"-"` // Typically from auth context, not part of stored doc directly if it's the doc ID

    FullName    string `json:"fullName" firestore:"fullName" binding:"required,min=2,max=100"`
    Email       string `json:"email" firestore:"email" binding:"required,email"`
    Phone       string `json:"phone,omitempty" firestore:"phone,omitempty" binding:"omitempty,max=20,custom_phone"` // Custom validation might be needed for specific regex
    BirthDate   string `json:"birthDate,omitempty" firestore:"birthDate,omitempty" binding:"omitempty,datetime=2006-01-02"`
    CEP         string `json:"cep,omitempty" firestore:"cep,omitempty" binding:"omitempty,max=9,custom_cep"`       // Custom validation for CEP
    City        string `json:"city,omitempty" firestore:"city,omitempty" binding:"omitempty,max=100"`
    State       string `json:"state,omitempty" firestore:"state,omitempty" binding:"omitempty,min=2,max=2"`
    // CreatedAt   time.Time `json:"createdAt,omitempty" firestore:"createdAt,omitempty"` // Managed by server
    // UpdatedAt   time.Time `json:"updatedAt,omitempty" firestore:"updatedAt,omitempty"` // Managed by server
}
