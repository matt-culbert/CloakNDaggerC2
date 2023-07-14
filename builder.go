// Original code from https://levelup.gitconnected.com/writing-a-code-generator-in-go-420e69151ab1

package main

import (
	"embed"
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
	Pubkey   string
}

func main() {
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

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

	// execute python script to generate keys
	genKey := exec.Command("python3", "crypter.py", uuid)
	out, err := genKey.Output()
	if err != nil {
		log.Fatal(err)
	}

	// read key
	pubPEM, err := ioutil.ReadFile("keys/" + uuid + ".pub.pem")
	if err != nil {
		panic(err)
	}

	string_pem := string(pubPEM)
	string_pem_no_newLines := strings.Replace(string_pem, "\n", "", -1)
	// Here we need to trim the start and end from the string
	string_pem_no_newLines = string_pem_no_newLines[:len(string_pem_no_newLines)-24]
	string_pem_no_newLines = string_pem_no_newLines[26:len(string_pem_no_newLines)]

	values.CallBack = os.Args[4]
	values.AppName = os.Args[3]
	values.UUID = uuid
	values.Pubkey = string_pem_no_newLines

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
	out, err = setEnvVar.Output()
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
