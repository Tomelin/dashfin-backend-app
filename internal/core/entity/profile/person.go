package entity_profile

import (
	"fmt"
	"time"
)

type ContractTypeValues string

const (
	ContractTypeCTL           ContractTypeValues = "clt"
	ContractTypePJ            ContractTypeValues = "pj"
	ContractTypeTemporary     ContractTypeValues = "temporary"
	ContractTypeInternship    ContractTypeValues = "internship"
	ContractTypePublicServant ContractTypeValues = "public_servant"
	ContractTypeOther         ContractTypeValues = "other"
)

type ProfileProfession struct {
	Profession    string             `json:"profession" binding:"required"`
	Company       string             `json:"company,omitempty"`
	ContractType  ContractTypeValues `json:"contractType"`
	MonthlyIncome float64            `json:"monthlyIncome,omitempty"`
}

type Goals struct {
	Name         string    `json:"name" binding:"required"`
	TargetDate   string    `json:"targetDate,omitempty"`
	Description  string    `json:"description,omitempty"`
	TargetAmount float64   `json:"targetAmount,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
}

type ProfileGoals struct {
	Goals2Years  []Goals   `json:"goals2Years" `
	Goals5Years  []Goals   `json:"goals5Years" `
	Goals10Years []Goals   `json:"goals10Years" `
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
}

type Profile struct {
	ID             string            `json:"id,omitempty"`
	FirstName      string            `json:"firstName"`
	LastName       string            `json:"lastName"`
	Email          string            `json:"email"`
	Phone          string            `json:"phone,omitempty"`
	BirthDate      string            `json:"birthDate,omitempty"`
	Sexo           string            `json:"sexo"`
	Cep            string            `json:"cep,omitempty"`
	City           string            `json:"city,omitempty"`
	State          string            `json:"state,omitempty"`
	UserProviderID string            `json:"userProviderID,omitempty"`
	Profession     ProfileProfession `json:"profession,omitempty"`
	Goals          ProfileGoals      `json:"goals,omitempty"`
	CreatedAt      time.Time         `json:"createdAt,omitempty"`
	UpdatedAt      time.Time         `json:"updatedAt,omitempty"`
}

func NewProfilePerson() {}

func (p *Profile) Validate() error {
	return nil
}

func (p *Profile) ValidateBirthDate(dateString string) (time.Time, error) {

	if dateString == "" {
		return time.Time{}, fmt.Errorf("birth_date is required")
	}

	layout := "2006-01-02"
	parsedTime, err := time.Parse(layout, dateString)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %w", err)
	}

	return parsedTime, nil
}

func (p *Profile) SetBirthDate(birth string) error {
	dateString, err := p.ValidateBirthDate(birth)
	if err != nil {
		return err
	}
	p.BirthDate = dateString.Format("2006-01-02")
	return nil
}
