package shared

import (
	"NULL/0x27894365/base_config/rogue"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

// SecureStruct holds the mutex, seed, and Counter.
// The mutex allows it to be locked, the seed is set at launch time, and the Counter is incremented every call
// The Counter can be changed to an algorithmically complicated increment
// The seed can be changed to a random string that is changed every launch cycle
type SecureStruct struct {
	mu      sync.Mutex
	seed    string
	Counter int
}

// NewSecureFunction creates a new SecureStruct with the given seed.
func NewSecureFunction(seed string) *SecureStruct {
	return &SecureStruct{
		seed:    seed,
		Counter: 0,
	}
}

// DeriveCount derives the current counter value after incrememnting it
func (s *SecureStruct) DeriveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Counter++
	return s.Counter
}

// Private method to generate the key for the current Counter.
func (s *SecureStruct) generateKey() string {
	// Derive a key using the seed and Counter
	data := fmt.Sprintf("%s:%d", s.seed, s.Counter)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ProtectedCaller allows an external caller to generate the key for a given Counter.
func ProtectedCaller(seed string, counter int) string {
	data := fmt.Sprintf("%s:%d", seed, counter)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// DecodeIPv6ToString takes a string formatted like an IPv6 address (e.g., ["7465:7374::"])
// and decodes it back into its original string representation.
func DecodeIPv6ToString(encoded string) (string, error) {
	// Remove all non-hexadecimal characters from the encoded string
	hexParts := strings.ReplaceAll(encoded, `"`, "")
	hexParts = strings.ReplaceAll(hexParts, `[`, "")
	hexParts = strings.ReplaceAll(hexParts, `]`, "")
	hexParts = strings.ReplaceAll(hexParts, `:`, "")
	hexParts = strings.ReplaceAll(hexParts, `,`, "")

	// Remove spaces and any non-hex characters explicitly
	hexParts = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || !unicode.Is(unicode.Hex_Digit, r) {
			return -1 // Skip non-hex characters
		}
		return r
	}, hexParts)

	// Pad the hex string if its length is odd
	if len(hexParts)%2 != 0 {
		hexParts = "0" + hexParts
	}

	// Decode the hex string into bytes
	result, err := hex.DecodeString(hexParts)
	if err != nil {
		return "", err
	}

	// Return the decoded string
	return string(result), nil
}

// VerifyMessageWithHMAC takes in a message, a received HMAC, and a secret key
// and returns true if the HMAC is valid for the message and secret key.
func (s *SecureStruct) VerifyMessageWithHMAC(message, receivedHMAC, providedKey string, secretKey []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate the expected key for this call
	expectedKey := s.generateKey()
	if providedKey != expectedKey {
		// The provided key is incorrect so don't proceed
		rogue.FuncDF7858354()
		return false
	}
	fmt.Printf("keys match %s %s\n", expectedKey, providedKey)
	// Create a new HMAC object using SHA-256
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(message))
	computedHMAC := h.Sum(nil)

	// Decode the received HMAC from hexadecimal
	receivedHMACBytes, err := hex.DecodeString(receivedHMAC)
	if err != nil {
		fmt.Println("Error decoding HMAC:", err)
		return false
	}

	// Use hmac.Equal to securely compare the computed HMAC with the received HMAC
	return hmac.Equal(computedHMAC, receivedHMACBytes)
}

// GenerateAuthToken takes in a URI and PSK
// and returns an auth token to use when requesting commands
// Update to add secret key requirement before producing the auth token
func (s *SecureStruct) GenerateAuthToken(uri, psk, providedKey string) (string, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate the expected key for this call
	expectedKey := s.generateKey()
	if providedKey != expectedKey {
		// The provided key is incorrect so don't proceed
		rogue.FuncDF7858354()
		return "", ""
	}
	// Generate a timestamp (current Unix time as a string)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Create the message by combining URI and timestamp
	message := uri + ":" + timestamp

	// Generate HMAC using SHA-256 and the secret key
	h := hmac.New(sha256.New, []byte(psk))
	h.Write([]byte(message))
	token := hex.EncodeToString(h.Sum(nil))

	return token, timestamp
}

// GetSelfHash gets the hash of itself during run time
func GetSelfHash() []byte {
	// Get the path of the currently running executable
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	// Open the executable file
	exeFile, err := os.Open(exePath)
	if err != nil {
		log.Fatalf("Failed to open executable file: %v", err)
	}
	defer exeFile.Close()

	// Create a SHA-256 hasher and copy the file's contents into it
	hasher := sha256.New()
	if _, err := io.Copy(hasher, exeFile); err != nil {
		log.Fatalf("Failed to hash executable file: %v", err)
	}
	return hasher.Sum(nil)
}
