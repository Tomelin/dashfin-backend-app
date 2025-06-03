package cryptdata

import (
	"errors"

	"encoding/base64"
)

type CryptDataInterface interface {
	Encode(data byte) (string, error)
	Decode(data *string) (byte, error)
}

type CryptData struct {
	Payload     string `json:"payload" binding:"required"`
	token       string
	tokenBase64 string
}

func InicializationCryptData(token *string) (CryptDataInterface, error) {

	if token == nil || *token == "" {
		return nil, errors.New("token is nil")
	}

	data := &CryptData{
		token: *token,
	}

	data.tokenEncode()

	return data, nil

}

func (c *CryptData) Encode(data byte) (string, error) {

	return "", nil

}

func (c *CryptData) Decode(data *string) (byte, error) {}

func (c *CryptData) tokenEncode() {

	c.tokenBase64 = base64.StdEncoding.EncodeToString([]byte(c.token))
}
