package utils

import (
	"encoding/json"
	"log"
	"os"
)

type settings struct {
	Database []struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		Id       string `json:"id"`
		Password string `json:"pw"`
	} `json:"database"`
}

var Setting settings

func init() {
	Setting = ReadSettings()
}

func ReadSettings() settings {
	file, err := os.OpenFile("../settings.json", os.O_RDONLY, os.FileMode(0644))
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(file)
	var s settings
	decoder.Decode(&s)
	return s
}
