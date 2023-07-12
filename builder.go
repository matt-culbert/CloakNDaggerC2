// Original code from https://levelup.gitconnected.com/writing-a-code-generator-in-go-420e69151ab1

package main

import (
	"crypto/rand"
	"crypto/rsa"
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

	"github.com/manifoldco/promptui"
)

var (
	//go:embed templates/*.tmpl
	rootFs embed.FS
)

type appValues struct {
	CallBack string
	AppName  string
	UUID     string
	pubkey   []byte
}

func main() {
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	filename := uuid
	bitSize := 4096

	if len(os.Args) < 5 {
		fmt.Printf("Not enough arguments. Need platform, architecture, output file name, and callback URL \n")
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

	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		panic(err)
	}

	// Extract public component.
	pub := key.Public()

	// Encode private key to PKCS#1 ASN.1 PEM.
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	// Encode public key to PKCS#1 ASN.1 PEM.
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		},
	)

	// Write private key to file.
	if err := ioutil.WriteFile("keys/"+filename+".pem", keyPEM, 0700); err != nil {
		panic(err)
	}

	// Write public key to file.
	if err := ioutil.WriteFile("keys/"+filename+".pub.pem", pubPEM, 0755); err != nil {
		panic(err)
	}

	values.CallBack = os.Args[4]
	values.AppName = os.Args[3]
	values.UUID = uuid
	values.pubkey = pubPEM

	rootFsMapping := map[string]string{
		"dagger.go.tmpl": mydir + "/templates/" + values.AppName + ".go",
	}

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

	fmt.Printf(" Done")

	// We set the app name and full path here for use later
	appNamePath := mydir + "/templates/" + values.AppName + ".go"

	// we set these as global compile options
	os.Setenv("GOOS", os.Args[1])
	os.Setenv("GOARCH", os.Args[2])

	// after setting environment variables, we compile using go build and the path to the file
	setEnvVar := exec.Command("go", "build", appNamePath)
	out, err := setEnvVar.Output()
	if err != nil {
		log.Fatal(err)
	}
	res := string(out)
	fmt.Printf(res)
}

func stringPrompt(label, defaultValue string) string {
	var (
		err    error
		result string
	)

	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
	}

	if result, err = prompt.Run(); err != nil {
		log.Fatalln(err)
	}

	return result
}
