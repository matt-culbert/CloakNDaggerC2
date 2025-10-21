//go:build withComp

package shared

import (
	"bytes"
	"compress/zlib"
	_ "embed"
	"encoding/json"
	"io"
	"log"
)

// Byte array which holds the config file embedded at compile time
//
//go:embed config.bin
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

func DoComp(data string) (bytes.Buffer, bool) {
	// Create a buffer to hold the compressed data
	var compressedData bytes.Buffer

	// Create a new zlib writer with the buffer as the output
	writer := zlib.NewWriter(&compressedData)

	// Compress the data
	_, err := writer.Write([]byte(data))
	if err != nil {
		// Return uncompressed data in case of an error
		return *bytes.NewBufferString(data), false
	}

	// Close the writer to finalize compression
	err = writer.Close()
	if err != nil {
		return bytes.Buffer{}, false
	}

	// Return the compressed data
	return compressedData, true
}

// Function to decompress zlib data and extract JSON fields
func LoadConfig() (Config, error) {
	var config Config

	// Create a new reader from the input buffer
	b := bytes.NewReader(configData)

	// Create a zlib reader to decompress the data
	r, err := zlib.NewReader(b)
	if err != nil {
		log.Println("Failed to create zlib reader:", err)
		return config, err
	}
	defer func(r io.ReadCloser) {
		err := r.Close()
		if err != nil {
			panic(err)
		}
	}(r) // Close the reader to avoid resource leaks

	// Read the decompressed data into a buffer
	var out bytes.Buffer
	_, err = io.Copy(&out, r)
	if err != nil {
		log.Println("Failed to decompress data:", err)
		return config, err
	}

	// Parse the JSON from the decompressed data
	err = json.Unmarshal(out.Bytes(), &config)
	if err != nil {
		log.Println("Failed to unmarshal JSON:", err)
		return config, err
	}
	return config, nil
}
