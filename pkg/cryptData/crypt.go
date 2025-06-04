package cryptdata

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64" // Usado para codificar/decodificar a chave e o payload.

	// "encoding/hex" // Não é mais necessário com a correção em EncryptPayload.
	// "encoding/json" // Não é mais necessário pois EncryptPayload agora recebe []byte.
	"errors"
	"fmt"
	"io"
)

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

// PayloadData é uma função wrapper que descriptografa usando a chave global do pacote.
func (c *CryptData) PayloadData(base64Payload string) ([]byte, error) {
	decryptedData, err := c.DecryptPayload(base64Payload, base64Key)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err) // Usar %w para wrapping de erro
	}
	return decryptedData, nil // Retornar nil para o erro em caso de sucesso
}

// pkcs7Unpad remove o padding PKCS7 dos dados.
func (c *CryptData) pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("pkcs7Unpad: input data is empty")
	}

	paddingLen := int(data[len(data)-1])

	if paddingLen == 0 || paddingLen > blockSize || paddingLen > len(data) {
		return nil, errors.New("pkcs7Unpad: invalid padding length (possible wrong key or corrupted data)")
	}

	// Opcional: Validar se todos os bytes de padding são iguais a paddingLen
	// for i := 0; i < paddingLen; i++ {
	//    if data[len(data)-1-i] != byte(paddingLen) {
	//        return nil, errors.New("pkcs7Unpad: invalid padding bytes")
	//    }
	// }

	return data[:len(data)-paddingLen], nil
}

// DecryptPayload descriptografa um payload que foi criptografado usando AES-CBC.
// Espera-se que base64Payload seja Base64(bytes_crus_IV + bytes_crus_Ciphertext).
func (c *CryptData) DecryptPayload(base64Payload string, base64KeyInput string) ([]byte, error) { // Renomeado parâmetro para evitar confusão com a constante
	if base64Payload == "" {
		return nil, errors.New("decrypt: base64 payload is empty")
	}
	if base64KeyInput == "" {
		return nil, errors.New("decrypt: base64 key is empty")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(base64KeyInput)
	if err != nil {
		return nil, fmt.Errorf("decrypt: failed to decode base64 key: %w", err)
	}

	switch len(keyBytes) {
	case 16, 24, 32:
	default:
		return nil, fmt.Errorf("decrypt: invalid AES key size: %d bytes (must be 16, 24, or 32)", len(keyBytes))
	}

	combinedBytes, err := base64.StdEncoding.DecodeString(base64Payload)
	if err != nil {
		return nil, fmt.Errorf("decrypt: failed to decode base64 payload: %w", err)
	}

	if len(combinedBytes) < aes.BlockSize {
		return nil, fmt.Errorf("decrypt: combined payload too short to contain IV (got %d bytes, expected at least %d)", len(combinedBytes), aes.BlockSize)
	}

	iv := combinedBytes[:aes.BlockSize]
	ciphertext := combinedBytes[aes.BlockSize:]

	if len(ciphertext) == 0 {
		return nil, errors.New("decrypt: ciphertext is empty after IV extraction")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("decrypt: ciphertext length (%d) is not a multiple of AES block size (%d)", len(ciphertext), aes.BlockSize)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("decrypt: failed to create AES cipher: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext) // Descriptografa in-place

	unpaddedData, err := c.pkcs7Unpad(ciphertext, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("decrypt: failed to unpad data: %w", err)
	}

	return unpaddedData, nil
}

// pkcs7Pad adiciona padding PKCS7 aos dados.
func (c *CryptData)  pkcs7Pad(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 || blockSize > 255 { // blockSize > 255 porque o byte de padding não pode exceder 255
		return nil, fmt.Errorf("pkcs7Pad: invalid block size %d (must be > 0 and <= 255)", blockSize)
	}
	padding := blockSize - (len(data) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...), nil
}

// EncryptPayload criptografa os jsonDataBytes fornecidos (que devem ser dados JSON já serializados)
// usando AES-CBC. O output é uma string Base64 no formato: Base64(bytes_crus_IV + bytes_crus_Ciphertext).
// Utiliza a chave global 'base64Key' definida no pacote.
func (c *CryptData)  EncryptPayload(jsonDataBytes []byte) (string, error) {
	if base64Key == "" { // Verifica a constante global
		return "", errors.New("encrypt: package key (base64Key) is empty")
	}

	// 1. Decodificar a chave de Base64 para bytes (usando a constante do pacote)
	keyBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		// Este erro seria inesperado se a constante base64Key for válida
		return "", fmt.Errorf("encrypt: failed to decode package base64 key: %w", err)
	}

	// Validar o tamanho da chave
	switch len(keyBytes) {
	case 16, 24, 32:
	default:
		// Este erro seria inesperado se a constante base64Key for válida
		return "", fmt.Errorf("encrypt: invalid package AES key size: %d bytes (must be 16, 24, or 32)", len(keyBytes))
	}

	// 2. Os jsonDataBytes já são fornecidos, não precisamos fazer json.Marshal aqui.

	// 3. Aplicar Padding PKCS7 aos dados JSON
	paddedData, err := c.pkcs7Pad(jsonDataBytes, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("encrypt: failed to apply PKCS7 padding: %w", err)
	}

	// 4. Criar o cipher AES
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("encrypt: failed to create AES cipher: %w", err)
	}

	// 5. Gerar um IV aleatório
	iv := make([]byte, aes.BlockSize) // aes.BlockSize é 16 bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("encrypt: failed to generate IV: %w", err)
	}

	// 6. Criptografar os dados usando o modo CBC
	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	// --- INÍCIO DAS ALTERAÇÕES CRÍTICAS ---
	// 7. Combinar IV (bytes crus) e Ciphertext (bytes crus)
	combinedRawBytes := append(iv, ciphertext...)

	// 8. Codificar os bytes combinados para Base64
	// Este formato é Base64(bytes_crus_IV + bytes_crus_Ciphertext),
	// que é o que DecryptPayload espera e o que o frontend (CryptoJS) originalmente produzia.
	base64EncryptedPayload := base64.StdEncoding.EncodeToString(combinedRawBytes)
	// --- FIM DAS ALTERAÇÕES CRÍTICAS ---

	return base64EncryptedPayload, nil
}
