package cryptdata

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"log"

	"encoding/base64"
)

type CryptDataInterface interface {
	DecodePayload(payload *string) ([]byte, error)
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

func (c *CryptData) DecodePayload(payload *string) ([]byte, error) {

	decodedPayload, err := base64.StdEncoding.DecodeString(*payload)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode payload: %w", err)
	}

	// Assuming the token is the key and IV concatenated.
	// In a real application, you would parse the key and IV from the token
	// based on your token format.
	keyAndIV, err := base64.StdEncoding.DecodeString(*c.token)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode token for key/IV: %w", err)
	}

	if len(keyAndIV) < 32 {
		return nil, errors.New("token does not contain sufficient data for key and IV")
	}
	key := keyAndIV[:16]  // Assuming a 16-byte key (AES-128)
	iv := keyAndIV[16:32] // Assuming a 16-byte IV

	stringToken, err := c.getTokenToString()
	log.Println(err)
	if err != nil {
		return nil, err
	}

	if payload == nil || *payload == "" {
		log.Println(err)
		return nil, errors.New("payload is nil")
	}

	if stringToken == nil || *stringToken == "" {
		log.Println(err)
		return nil, errors.New("token is nil")
	}

	if err := c.validateToken(); err != nil {
		log.Println(err)
		return nil, err
	}

	block, err := aes.NewCipher(key)
	log.Println(err)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Use CBC mode for decryption with the provided IV.
	// You will need to handle unpadding after decryption based on the padding scheme used during encryption.
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(decodedPayload))
	mode.CryptBlocks(decrypted, decodedPayload)
	log.Println(string(decrypted))
	// In a real application, you would unpad the decryptedPayload here.
	return decrypted, nil
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
