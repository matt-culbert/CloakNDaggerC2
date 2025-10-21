//go:build !withComp

package shared

import (
	"bytes"
	_ "embed"
	"encoding/json"
)

// Byte array which holds the config file embedded at compile time
//
//go:embed config.json
var configData []byte // Embedded config data

// Config struct to hold configuration
type Config struct {
	Method   string `json:"method"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	GetPath  string `json:"get_path"`
	PostPath string `json:"post_path"`
	Id       string `json:"id"`
	Sleep    string `json:"sleep"`
	Psk1     string `json:"psk1"`
	Psk2     string `json:"psk2"`
}

// Function to load configuration from embedded JSON
func LoadConfig() (Config, error) {
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func DoComp(data string) (bytes.Buffer, bool) {
	tossAway := bytes.NewBufferString(data)
	return *tossAway, false

}
