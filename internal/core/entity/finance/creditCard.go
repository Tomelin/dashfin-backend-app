package entity_finance

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

type CreditCardRepositoryInterface interface {
	CreateCreditCard(ctx context.Context, data *CreditCard) (*CreditCardRequest, error)
	GetCreditCardByID(ctx context.Context, id *string) (*CreditCardRequest, error)
	GetCreditCards(ctx context.Context) ([]CreditCardRequest, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]CreditCardRequest, error)
	UpdateCreditCard(ctx context.Context, data *CreditCardRequest) (*CreditCardRequest, error)
	DeleteCreditCard(ctx context.Context, id *string) error
}

type CreditCardServiceInterface interface {
	CreateCreditCard(ctx context.Context, data *CreditCard) (*CreditCardRequest, error)
	GetCreditCardByID(ctx context.Context, id *string) (*CreditCardRequest, error)
	GetCreditCards(ctx context.Context) ([]CreditCardRequest, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]CreditCardRequest, error)
	UpdateCreditCard(ctx context.Context, data *CreditCardRequest) (*CreditCardRequest, error)
	DeleteCreditCard(ctx context.Context, id *string) error
}

type CreditCard struct {
	CardBrand       string  `json:"cardBrand" bson:"cardBrand"`
	CustomCardBrand string  `json:"customCardBrand,omitempty" bson:"customCardBrand,omitempty"`
	Description     string  `json:"description,omitempty" bson:"description,omitempty"`
	LastFourDigits  string  `json:"lastFourDigits" bson:"lastFourDigits"`
	InvoiceDueDate  int     `json:"invoiceDueDate" bson:"invoiceDueDate"`
	CardExpiryMonth int     `json:"cardExpiryMonth" bson:"cardExpiryMonth"`
	CardExpiryYear  int     `json:"cardExpiryYear" bson:"cardExpiryYear"`
	CreditLimit     float64 `json:"creditLimit" bson:"creditLimit"`
}

type CreditCardRequest struct {
	ID string `json:"id"`
	CreditCard
}

type CreditCardResponse CreditCardRequest

func NewCreditCard(cardBrand, lastFourDigits string, invoiceDueDate, cardExpiryMonth, cardExpiryYear int, creditLimit float64) *CreditCard {
	return &CreditCard{
		CardBrand:       cardBrand,
		LastFourDigits:  lastFourDigits,
		InvoiceDueDate:  invoiceDueDate,
		CardExpiryMonth: cardExpiryMonth,
		CardExpiryYear:  cardExpiryYear,
		CreditLimit:     creditLimit,
	}
}

func (cc *CreditCard) Validate() error {
	if strings.TrimSpace(cc.CardBrand) == "" {
		return errors.New("cardBrand is required")
	}

	validBrands := map[string]bool{
		"mastercard": true,
		"visa":       true,
		"elo":        true,
		"amex":       true,
		"hipercard":  true,
		"diners":     true,
		"discover":   true,
		"jcb":        true,
		"aura":       true,
		"other":      true,
	}

	if !validBrands[cc.CardBrand] {
		return errors.New("cardBrand must be one of: mastercard, visa, elo, amex, hipercard, diners, discover, jcb, aura, other")
	}

	if cc.CardBrand == "other" && strings.TrimSpace(cc.CustomCardBrand) == "" {
		return errors.New("customCardBrand is required when cardBrand is 'other'")
	}

	if len(cc.CustomCardBrand) > 50 {
		return errors.New("customCardBrand must not exceed 50 characters")
	}

	if len(cc.Description) > 150 {
		return errors.New("description must not exceed 150 characters")
	}

	lastFourDigits := strings.TrimSpace(cc.LastFourDigits)
	if lastFourDigits == "" {
		return errors.New("lastFourDigits is required")
	}
	if len(lastFourDigits) != 4 {
		return errors.New("lastFourDigits must be exactly 4 characters")
	}
	matched, _ := regexp.MatchString(`^\d{4}$`, lastFourDigits)
	if !matched {
		return errors.New("lastFourDigits must contain only numbers")
	}

	if cc.InvoiceDueDate < 1 || cc.InvoiceDueDate > 31 {
		return errors.New("invoiceDueDate must be between 1 and 31")
	}

	if cc.CardExpiryMonth < 1 || cc.CardExpiryMonth > 12 {
		return errors.New("cardExpiryMonth must be between 1 and 12")
	}

	if cc.CardExpiryYear < 1000 || cc.CardExpiryYear > 9999 {
		return errors.New("cardExpiryYear must be a 4-digit year")
	}

	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	if cc.CardExpiryYear < currentYear {
		return errors.New("cardExpiryYear must be current year or later")
	}

	if cc.CardExpiryYear == currentYear && cc.CardExpiryMonth < currentMonth {
		return errors.New("card expiry date cannot be in the past")
	}

	if cc.CreditLimit < 0 {
		return errors.New("creditLimit must be greater than or equal to 0")
	}

	return nil
}
