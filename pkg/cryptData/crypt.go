package cryptdata

import (
	"errors"
	"fmt"
	"log"

	"encoding/base64"
)

type CryptDataInterface interface {
	// Encode(data byte) (string, error)
	// Decode(data *string) (byte, error)
	// DecryptPayload(payload string) (string, error)
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

// func (c *CryptData) Encode(data byte) (string, error) {

// 	return "", nil

// }

// func (c *CryptData) Decode(data *string) (byte, error) { return 0, nil }

// func (c *CryptData) DecryptPayload(payload string) (string, error) {
// 	decodedPayload, err := base64.StdEncoding.DecodeString(payload)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode base64 payload: %w", err)
// 	}

// 	err = c.getStringToken()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode base64 key: %w", err)
// 	}

// 	block, err := aes.NewCipher([]byte(*c.token))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create AES cipher: %w", err)
// 	}

// 	if len(decodedPayload)%aes.BlockSize != 0 {
// 		return "", errors.New("ciphertext is not a multiple of the block size")
// 	}

// 	decrypted := make([]byte, len(decodedPayload))

// 	for i := range decodedPayload {
// 		decrypted[i] = decodedPayload[i] ^ decodedKey[i%len(decodedKey)]
// 	}

// 	return string(decrypted), nil
// }

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
