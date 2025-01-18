package config

import (
	"encoding/json"
	"os"
)

type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`

	AssemblyName string `json:"assemblyname"`
	Assembly     string `json:"assembly"`
	Procedure    string `json:"procedure"`

	Functions []Function `json:"functions"`
}

var CONFIG Config

func Load(path string) error {
	byteValue, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteValue, &CONFIG)
}
