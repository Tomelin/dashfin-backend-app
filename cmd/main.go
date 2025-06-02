package main

import (
	"github.com/Tomelin/dashfin-backend-app/config"
)
func main(){

	cfg, err := config.LoadConfig()
	log.Println(err)
	log.Println(cfg)
}