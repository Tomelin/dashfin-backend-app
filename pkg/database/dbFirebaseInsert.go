package database

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type BrazilianBank struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type BanksData struct {
	Banks []BrazilianBank `json:"banks"`
}

type FirebaseInsertInterface interface {
	InsertBrazilianBanksFromJSON(ctx context.Context) error
}

type FirebaseInsert struct {
	firebaseDB FirebaseDBInterface
	collection string
	filePath   string
}

func NewFirebaseInsert(firebaseDB FirebaseDBInterface) FirebaseInsertInterface {
	return &FirebaseInsert{
		firebaseDB: firebaseDB,
		collection: "platform_financial-institution",
		filePath:   "/home/user/dashfin-backend-app/files/banks.json",
	}
}

func (fi *FirebaseInsert) InsertBrazilianBanksFromJSON(ctx context.Context) error {
	data, err := os.ReadFile(fi.filePath)
	if err != nil {
		return err
	}

	var banksData BanksData
	if err := json.Unmarshal(data, &banksData); err != nil {
		return err
	}

	for _, bank := range banksData.Banks {
		log.Println(bank)
		bankMap, err := utils.StructToMap(bank)
		if err != nil {
			return err
		}
		log.Println(bankMap)
		result, err := fi.firebaseDB.Create(ctx, bankMap, fi.collection)
		log.Println(" result ", result, "err", err)
		if err != nil {
			return err
		}
	}

	return nil
}
