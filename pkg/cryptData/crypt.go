package cryptdata

import (
	"errors"
	"fmt"
	"log"

	"encoding/base64"
)

type CryptDataInterface interface {
	GetToken()
}

type CryptData struct {
	Payload string `json:"payload" binding:"required"`
	token   *string
}

func InicializationCryptData(token *string) (CryptDataInterface, error) {

	if token == nil || *token == "" {
		return nil, errors.New("token is nil")
	}

	data := &CryptData{}
	err := data.validateTokenFromString(token)
	if err != nil {
		return nil, err
	}

	data.token = token

	return data, nil
}

func (c *CryptData) GetPayload() string {
	return c.Payload
}

func (c *CryptData) GetToken() {
	log.Println("token is....")
	log.Println(c.token)
	log.Println(*c.token)
}

func (c *CryptData) validateTokenFromString(token *string) error {
	decodeToken, err := base64.StdEncoding.DecodeString(*token)
	if err != nil {
		return fmt.Errorf("failed to decode base64 key: %w", err)
	}

	if decodeToken == nil || string(decodeToken) == "" {
		return errors.New("token is nil")
	}

	return nil
}

func (c *CryptData) validateToken() error {
	decodeToken, err := base64.StdEncoding.DecodeString(*c.token)
	if err != nil {
		return fmt.Errorf("failed to decode base64 key: %w", err)
	}

	if decodeToken == nil || string(decodeToken) == "" {
		return errors.New("token is nil")
	}

	return nil
}

func (c *CryptData) getTokenToString() (*string, error) {
	decodeToken, err := base64.StdEncoding.DecodeString(*c.token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key: %w", err)
	}

	if decodeToken == nil || string(decodeToken) == "" {
		return nil, errors.New("token is nil")
	}

	t := string(decodeToken)

	return &t, nil
}
