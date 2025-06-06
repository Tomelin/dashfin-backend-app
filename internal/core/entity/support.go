package entity

import "errors"

type Support struct {
	Category       string `json:"category"`
	Description    string `json:"description"`
	UserProviderID string `json:"userProviderID"`
}

type SupportResponse struct {
	ID string `json:"id"`
	Support
}

func (s *Support) Validate() error {
	if s.Category == "" {
		return errors.New("category is required")
	}
	if s.Description == "" {
		return errors.New("description is required")
	}
	if s.UserProviderID == "" {
		return errors.New("userID is required")
	}

	return nil
}

func (s *SupportResponse) Validate() error {

	if s.ID == "" {
		return errors.New("id is required")
	}

	if err := s.Support.Validate(); err != nil {
		return err
	}

	return nil
}
