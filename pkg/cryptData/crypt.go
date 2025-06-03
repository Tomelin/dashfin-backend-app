package cryptdata

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
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
	// Assuming encryptionKeyBytes is available globally or passed to the struct
	// This should hold the base64 decoded BACKEND_ENCRYPTION_KEY
	// Example placeholder:
	var encryptionKeyBytes []byte // You need to initialize this with your actual key
	// encryptionKeyBytes = base64.StdEncoding.DecodeString(os.Getenv("BACKEND_ENCRYPTION_KEY")) // Example of how to get it

	// For demonstration, using a placeholder. Replace with your actual key.
	encryptionKeyBytes, err := base64.StdEncoding.DecodeString(*c.token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode backend encryption key: %w", err)
	}

	if payload == nil || *payload == "" {
		return nil, errors.New("payload is nil")
	}

	// 1. Base64 decode the incoming payload string
	decodedPayload, err := base64.StdEncoding.DecodeString(*payload)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode payload: %w", err)
	}

	// 2. The base64 decoded payload is a hex string (IV + ciphertext). Decode it from hex.
	hexDecodedPayload := make([]byte, hex.DecodedLen(len(decodedPayload)))
	n, err := hex.Decode(hexDecodedPayload, decodedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to hex decode payload: %w", err)
	}
	hexDecodedPayload = hexDecodedPayload[:n] // Trim to actual decoded length

	if err := c.validateToken(); err != nil {
		log.Println(err)
		return nil, err
	}

	// 3. Separate the IV (first 16 bytes) and the ciphertext
	if len(hexDecodedPayload) < aes.BlockSize {
		return nil, errors.New("hex decoded payload is too short to contain IV")
	}
	iv := hexDecodedPayload[:aes.BlockSize]
	ciphertext := hexDecodedPayload[aes.BlockSize:]

	// 4. Create a new AES cipher using the backend key
	block, err := aes.NewCipher(encryptionKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Check if the ciphertext length is a multiple of the block size (required for CBC with padding)
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	// 5. Create a CBC decrypter
	mode := cipher.NewCBCDecrypter(block, iv)

	// 6. Decrypt the ciphertext
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	// 7. Unpad the decrypted data (PKCS7)
	unpadded, err := pkcs7Unpad2(decrypted, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad decrypted data: %w", err)
	}

	return unpadded, nil
}

// pkcs7Unpad removes PKCS7 padding from data
func pkcs7Unpad2(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, errors.New("invalid block size")
	}
	if len(data)%blockSize != 0 || len(data) == 0 {
		return nil, errors.New("invalid data length for PKCS7 unpadding")
	}
	padding := int(data[len(data)-1])

	// In a real application, you would unpad the decryptedPayload here.
	return data[:len(data)-padding], nil
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
