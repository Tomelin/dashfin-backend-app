package main

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

var (
	token  = "bXkxNmRpZ2l0SXZLZXkxMg=="
	mydata = "meus dados aqui"
	key    = "my32digitkey12345678901234567890"
	iv     = "my16digitIvKey12"
)

func main() {

	// encode token para base64

	// pegamos o token no formato de []byte
	tokenString := decode(token)
	fmt.Println("received token ", string(tokenString))

	payload,_ := json.Marshal(mydata)
	fmt.Println(payload)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(block)

}

func encode() string {
	sEnc := base64.StdEncoding.EncodeToString([]byte(iv))
	fmt.Println(sEnc)
	return sEnc
}

func decode(data string) []byte {
	sDec, _ := base64.StdEncoding.DecodeString(data)
	return sDec
}

func encryptData(data byte) {

}

func decryptData() {}
