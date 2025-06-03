package cryptdata

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
)

// Exemplo de como usar (não faz parte do package, apenas para demonstração):
func PayloadData(base64Payload string) {
	// Mesma chave do exemplo TypeScript (decodificada)
	// "ThisIsA16ByteKeyThisIsA16ByteIV" -> VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo=
	// Para AES-256, use uma chave de 32 bytes. Ex: "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f" (hex)
	// que em Base64 seria: AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=

	// Chave usada no TypeScript (VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo= -> "ThisIsA16ByteKey")
	// Esta chave tem 16 bytes, então é AES-128.
	// A string "ThisIsA16ByteKeyThisIsA16ByteIV" na verdade tem 32 bytes.
	// Vamos usar a chave do exemplo: VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo=
	// Decodificando: "ThisIsA16ByteKeyThisIsA16ByteIV" (32 bytes)
	const base64Key = "VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo=" // 32 bytes key

	// Payload de exemplo (simulando o que o TypeScript enviaria)
	// Suponha que o dado original seja: {"message":"hello world"}
	// IV (16 bytes, exemplo): 000102030405060708090a0b0c0d0e0f (hex)
	// Ciphertext (exemplo, após AES-CBC com a chave acima e PKCS7 padding):
	// Se o dado original é {"message":"hello world"}, que tem 24 bytes.
	// Com padding PKCS7 para bloco de 16, adiciona-se 8 bytes de valor 0x08. Total 32 bytes.
	// Exemplo de output do CryptoJS (IVhex + CiphertextHex) e depois Base64.
	// Este valor precisa ser gerado pelo código TypeScript com a chave exata.
	// Vou usar um payload que gerei com o código TS fornecido e a chave "VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo="
	// para o input: {"data":"test"}
	// IV (hex): 6d2d706a58687232434b334c7439354f
	// Ciphertext (hex): e0d6923f12756ff02442c5307b513631
	// CombinedHex: 6d2d706a58687232434b334c7439354fe0d6923f12756ff02442c5307b513631
	// Base64(CombinedHex): bS1wamFYaHIyQ0szTHRNOTVPL+DWkj8SdWP/JELFMHtRNjE=
	// const base64Payload = "bS1wamFYaHIyQ0szTHRNOTVPL+DWkj8SdWP/JELFMHtRNjE=" // payload for {"data":"test"}

	decryptedData, err := DecryptPayload(base64Payload, base64Key)
	if err != nil {
		log.Fatalf("Decryption failed2: %v", err)
	}

	fmt.Printf("Decrypted data: %s\n", string(decryptedData)) // Esperado: {"data":"test"}
}

// pkcs7Unpad remove o padding PKCS7 dos dados.
// O blockSize é geralmente aes.BlockSize.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("pkcs7Unpad: input data is empty")
	}
	if len(data)%blockSize != 0 {
		// Isso não deveria acontecer se a criptografia foi feita corretamente
		// e o ciphertext não foi corrompido.
		return nil, errors.New("pkcs7Unpad: input data is not a multiple of the block size")
	}

	paddingLen := int(data[len(data)-1])

	// Validação do tamanho do padding
	// O padding não pode ser zero nem maior que o tamanho do bloco.
	// Também não pode ser maior que o próprio comprimento dos dados.
	if paddingLen == 0 || paddingLen > blockSize || paddingLen > len(data) {
		// Este erro pode indicar uma chave de descriptografia incorreta,
		// dados corrompidos, ou um padding malformado.
		return nil, errors.New("pkcs7Unpad: invalid padding length or corrupted data")
	}

	// Verifica se todos os bytes de padding são consistentes (opcional, mas bom para robustez)
	// Comentado para seguir a implementação mais comum que apenas remove o número de bytes indicado.
	// Algumas implementações podem não garantir que todos os bytes de padding tenham o valor paddingLen.
	// A verificação mais importante é que paddingLen seja um valor sensível.
	/*
		for i := 0; i < paddingLen; i++ {
			if data[len(data)-1-i] != byte(paddingLen) {
				return nil, fmt.Errorf("pkcs7Unpad: invalid padding bytes at position %d", len(data)-1-i)
			}
		}
	*/

	return data[:len(data)-paddingLen], nil
}

// DecryptPayload descriptografa um payload que foi criptografado usando AES-CBC
// com um IV concatenado (IV+Ciphertext), onde o conjunto foi convertido para hexadecimal
// e depois para Base64. A chave também é fornecida em Base64.
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

	// 2. Decodificar o payload de Base64 para obter a string hexadecimal (IVhex + CiphertextHex)
	hexPayloadBytes, err := base64.StdEncoding.DecodeString(base64Payload)
	if err != nil {
		return nil, fmt.Errorf("decrypt: failed to decode base64 payload: %w", err)
	}
	hexPayload := string(hexPayloadBytes)

	// 3. Converter a string hexadecimal combinada para bytes
	combinedBytes, err := hex.DecodeString(hexPayload)
	if err != nil {
		return nil, fmt.Errorf("decrypt: failed to decode hex payload: %w", err)
	}

	// 4. Separação do IV e Ciphertext
	// O IV tem sempre 16 bytes (aes.BlockSize)
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

	// 5. Criação do Cipher AES
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		// Este erro já é verificado pela validação do tamanho da chave, mas é bom tê-lo por completude.
		return nil, fmt.Errorf("decrypt: failed to create AES cipher: %w", err)
	}

	// 6. Descriptografia CBC
	// O CBCDecrypter precisa de um IV do mesmo tamanho do bloco da cifra.
	// E o ciphertext deve ser um múltiplo do tamanho do bloco.
	mode := cipher.NewCBCDecrypter(block, iv)

	// Criamos uma cópia do ciphertext para descriptografar, ou podemos descriptografar in-place.
	// Para descriptografar in-place, o slice de output e input devem ser o mesmo.
	// Se não quisermos modificar o 'ciphertext' original (que é uma fatia de combinedBytes),
	// deveríamos fazer uma cópia:
	// decrypted := make([]byte, len(ciphertext))
	// mode.CryptBlocks(decrypted, ciphertext)
	// Ou, para modificar 'ciphertext' (que é uma fatia, então 'combinedBytes' será modificado):
	mode.CryptBlocks(ciphertext, ciphertext) // Descriptografa in-place
	decryptedWithPadding := ciphertext       // Agora 'ciphertext' contém os dados descriptografados com padding

	// 7. Remoção do Padding PKCS7
	unpaddedData, err := pkcs7Unpad(decryptedWithPadding, aes.BlockSize)
	if err != nil {
		// Um erro aqui frequentemente significa que a chave estava incorreta ou os dados foram corrompidos.
		return nil, fmt.Errorf("decrypt: failed to unpad data (possible wrong key or corrupted data): %w", err)
	}

	return unpaddedData, nil
}
