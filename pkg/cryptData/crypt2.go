package cryptdata

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64" // Ainda pode ser útil para debugging ou outras funções, mas não diretamente neste fluxo.
	"errors"
	"fmt"
	"log"
)

// Exemplo de como usar (não faz parte do package, apenas para demonstração):
// Esta função agora pode ser usada para testar.
func PayloadData(base64Payload string) {
	// Usar a mesma chave que foi usada para criptografar este payload específico.
	// A chave "VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo=" é "ThisIsA16ByteKeyThisIsA16ByteIV" (32 bytes)
	const base64Key = "VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo="

	decryptedData, err := DecryptPayload(base64Payload, base64Key)
	if err != nil {
		log.Fatalf("Decryption failed: %v", err) // Removido o "2" para consistência
	}

	fmt.Printf("Decrypted data: %s\n", string(decryptedData))
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

	// Opcional: verificar se todos os bytes de padding são iguais a paddingLen.
	// for i := 0; i < paddingLen; i++ {
	//    if data[len(data)-1-i] != byte(paddingLen) {
	//        return nil, errors.New("pkcs7Unpad: invalid padding bytes")
	//    }
	// }

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
