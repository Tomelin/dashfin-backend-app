package dto

type IncomeRecord struct {
	ID               string  `json:"id" bson:"_id,omitempty"`
	Category         string  `json:"category"`
	Description      *string `json:"description,omitempty"`
	BankAccountID    string  `json:"bankAccountId"`
	Amount           float64 `json:"amount"`
	ReceiptDate      string  `json:"receiptDate"` // ISO 8601 (YYYY-MM-DD)
	IsRecurring      bool    `json:"isRecurring"`
	RecurrenceCount  *int    `json:"recurrenceCount,omitempty" ` // Pointer to allow null
	RecurrenceNumber int     `json:"recurrenceNumber,omitempty"` // Pointer to allow null
	Observations     *string `json:"observations,omitempty"`
	UserID           string  `json:"userId,omitempty"` // To associate with a user
}
