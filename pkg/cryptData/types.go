package cryptdata

type CryptDataInternalInterface interface {
	validateTokenFromString(token *string) error
	pkcs7Unpad(data []byte, blockSize int) ([]byte, error)
	pkcs7Pad(data []byte, blockSize int) ([]byte, error)
}

type CryptDataInterface interface {
	PayloadData(base64Payload string) ([]byte, error)
	DecryptPayload(base64Payload string, base64KeyInput string) ([]byte, error)
	EncryptPayload(jsonDataBytes []byte) (string, error)
	CryptDataInternalInterface
}

// const base64Key = "VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo="
