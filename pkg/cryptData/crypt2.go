package cryptdata

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64" // Ainda pode ser útil para debugging ou outras funções, mas não diretamente neste fluxo.
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const base64Key = "VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo="

// Exemplo de como usar (não faz parte do package, apenas para demonstração):
// Esta função agora pode ser usada para testar.
func PayloadData(base64Payload string) ([]byte, error) {
	
	decryptedData, err := DecryptPayload(base64Payload, base64Key)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %v", err)
	}
	return decryptedData, err
}

// pkcs7Unpad remove o padding PKCS7 dos dados.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("pkcs7Unpad: input data is empty")
	}
	// A verificação de len(data)%blockSize != 0 pode ser muito rigorosa aqui,
	// pois a descriptografia CBC já garante isso. Se os dados não fossem múltiplos,
	// a etapa de CryptBlocks provavelmente já teria problemas ou o padding seria o bloco inteiro.
	// No entanto, mantê-la não prejudica, mas pode ser redundante.

	paddingLen := int(data[len(data)-1])

	if paddingLen == 0 || paddingLen > blockSize || paddingLen > len(data) {
		// Este erro é crucial e geralmente indica chave errada ou dados corrompidos.
		return nil, errors.New("pkcs7Unpad: invalid padding length (possible wrong key or corrupted data)")
	}

	return data[:len(data)-paddingLen], nil
}

// DecryptPayload descriptografa um payload que foi criptografado usando AES-CBC
func DecryptPayload(base64Payload string, base64Key string) ([]byte, error) {
	if base64Payload == "" {
		return nil, errors.New("decrypt: base64 payload is empty")
	}
	if base64Key == "" {
		return nil, errors.New("decrypt: base64 key is empty")
	}

	// 1. Decodificar a chave de Base64 para bytes
	keyBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("decrypt: failed to decode base64 key: %w", err)
	}

	// Validar o tamanho da chave (AES-128, AES-192, ou AES-256)
	switch len(keyBytes) {
	case 16, 24, 32:
		// Tamanho de chave válido
	default:
		return nil, fmt.Errorf("decrypt: invalid AES key size: %d bytes (must be 16, 24, or 32)", len(keyBytes))
	}

	// 2. Decodificar o payload de Base64 para obter os bytes combinados (IV + Ciphertext)
	// O resultado de DecodeString já são os bytes crus que representam IV + Ciphertext.
	combinedBytes, err := base64.StdEncoding.DecodeString(base64Payload)
	if err != nil {
		// Este erro ocorreria se base64Payload não fosse Base64 válido.
		// O erro que você viu (encoding/hex) acontecia na etapa seguinte.
		return nil, fmt.Errorf("decrypt: failed to decode base64 payload: %w", err)
	}

	// A ETAPA 3 ANTERIOR (hex.DecodeString(string(hexPayloadBytes))) FOI REMOVIDA
	// PORQUE combinedBytes JÁ SÃO OS BYTES CORRETOS.

	// 4. Separação do IV e Ciphertext
	// aes.BlockSize é uma constante igual a 16.
	if len(combinedBytes) < aes.BlockSize {
		return nil, fmt.Errorf("decrypt: combined payload too short to contain IV (got %d bytes, expected at least %d)", len(combinedBytes), aes.BlockSize)
	}

	iv := combinedBytes[:aes.BlockSize]
	ciphertext := combinedBytes[aes.BlockSize:]

	if len(ciphertext) == 0 {
		return nil, errors.New("decrypt: ciphertext is empty after IV extraction")
	}
	// O ciphertext para CBC deve ter um comprimento que seja múltiplo do tamanho do bloco.
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("decrypt: ciphertext length (%d) is not a multiple of AES block size (%d)", len(ciphertext), aes.BlockSize)
	}

	// 5. Criação do Cipher AES
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		// Redundante se a validação de tamanho da chave acima for mantida, mas seguro ter.
		return nil, fmt.Errorf("decrypt: failed to create AES cipher: %w", err)
	}

	// 6. Descriptografia CBC
	// NewCBCDecrypter retorna um BlockMode.
	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks descriptografa no local (in-place).
	// O slice 'ciphertext' será modificado para conter os dados descriptografados com padding.
	mode.CryptBlocks(ciphertext, ciphertext)
	decryptedWithPadding := ciphertext // Apenas para clareza do nome da variável

	// 7. Remoção do Padding PKCS7
	unpaddedData, err := pkcs7Unpad(decryptedWithPadding, aes.BlockSize)
	if err != nil {
		// Um erro aqui frequentemente significa que a chave estava incorreta ou os dados foram corrompidos.
		return nil, fmt.Errorf("decrypt: failed to unpad data: %w", err)
	}

	return unpaddedData, nil
}

// pkcs7Pad adiciona padding PKCS7 aos dados.
func pkcs7Pad(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, fmt.Errorf("pkcs7Pad: invalid block size %d", blockSize)
	}
	padding := blockSize - (len(data) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...), nil
}

// EncryptPayload criptografa os dados fornecidos (que serão primeiro convertidos para JSON)
// usando AES-CBC. O output é uma string Base64 no formato: Base64(IVhex + CiphertextHex).
// A chave (base64Key) também é fornecida em Base64.
func EncryptPayload(dataToEncrypt interface{}) (string, error) {
	if base64Key == "" {
		return "", errors.New("encrypt: base64 key is empty")
	}

	// 1. Decodificar a chave de Base64 para bytes
	keyBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return "", fmt.Errorf("encrypt: failed to decode base64 key: %w", err)
	}

	// Validar o tamanho da chave (AES-128, AES-192, ou AES-256)
	switch len(keyBytes) {
	case 16, 24, 32:
		// Tamanho de chave válido
	default:
		return "", fmt.Errorf("encrypt: invalid AES key size: %d bytes (must be 16, 24, or 32)", len(keyBytes))
	}

	// 2. Converter os dados para JSON
	jsonDataBytes, err := json.Marshal(dataToEncrypt)
	if err != nil {
		return "", fmt.Errorf("encrypt: failed to marshal data to JSON: %w", err)
	}

	// 3. Aplicar Padding PKCS7 aos dados JSON
	// O modo CBC já lida com padding se o tamanho dos dados não for múltiplo do bloco,
	// mas é uma boa prática ser explícito com PKCS7, pois CryptoJS faz isso.
	paddedData, err := pkcs7Pad(jsonDataBytes, aes.BlockSize)
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

	// 7. Converter IV e Ciphertext para strings hexadecimais
	ivHex := hex.EncodeToString(iv)
	ciphertextHex := hex.EncodeToString(ciphertext)

	// 8. Concatenar IVhex e CiphertextHex
	combinedHex := ivHex + ciphertextHex

	// 9. Codificar a string hexadecimal combinada para Base64
	// O frontend espera Base64(HexToString(IV) + HexToString(Ciphertext))
	// Então, convertemos a string `combinedHex` para bytes antes de codificar para Base64.
	base64EncryptedPayload := base64.StdEncoding.EncodeToString([]byte(combinedHex))

	return base64EncryptedPayload, nil
}
