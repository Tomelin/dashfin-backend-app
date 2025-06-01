package support

// SupportRequestCategory defines the type for support request categories.
type SupportRequestCategory string

// Defines the allowed enum values for support request categories.
const (
	CategoryTechnicalSupport      SupportRequestCategory = "technical_support"
	CategoryNewFeatureSuggestion  SupportRequestCategory = "new_feature_suggestion"
	CategoryBillingIssue          SupportRequestCategory = "billing_issue"
	CategoryGeneralQuestion       SupportRequestCategory = "general_question"
	CategoryOther                 SupportRequestCategory = "other"
)

// IsValidCategory checks if the provided category is valid.
func (c SupportRequestCategory) IsValid() bool {
	switch c {
	case CategoryTechnicalSupport, CategoryNewFeatureSuggestion, CategoryBillingIssue, CategoryGeneralQuestion, CategoryOther:
		return true
	}
	return false
}
