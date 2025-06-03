package entity_profile

import (
	"fmt"
	"time"
)

type Profile struct {
	FullName  string `json:"fullName"` 
	Email     string `json:"email"`    
	Phone     string `json:"phone,omitempty"`
	BirthDate string `json:"birthDate,omitempty"`
	Sexo      string `json:"sexo"` // Correto!
	Cep       string `json:"cep,omitempty"`
	City      string `json:"city,omitempty"`
	State     string `json:"state,omitempty"`
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
