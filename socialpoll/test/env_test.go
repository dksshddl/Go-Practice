package test

import (
	"encoding/json"
	"os"
	"testing"
)

type settings struct {
	Database []struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		Id       string `json:"id"`
		Password string `json:"pw"`
	} `json:"database"`
}

func TestEnv(t *testing.T) {
	value := os.Getenv("SP_TWITTER_BEARERTOKEN")
	t.Log("TWITTER BEARERTOKEN : ", value)
	if value == "" {
		t.Error("cannot find environment variable")
	}
}

func TestSetting(t *testing.T) {
	file, err := os.OpenFile("../../settings.json", os.O_RDONLY, os.FileMode(0644))
	if err != nil {
		t.Error(err)
	}
	decoder := json.NewDecoder(file)
	var s settings
	decoder.Decode(&s)
	dir, _ := os.Getwd()
	t.Log(dir)
	t.Log(s.Database[0])
	t.Fail()
}
