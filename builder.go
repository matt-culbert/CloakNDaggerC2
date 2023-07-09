// Original code from https://levelup.gitconnected.com/writing-a-code-generator-in-go-420e69151ab1

package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"uuid"

	"github.com/manifoldco/promptui"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed templates/*.tmpl
	rootFs embed.FS
)

type appValues struct {
	CallBack string
	AppName  string
	UUID     string
	pubkey   string
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	ctx := context.Background()

	err := client.Set(ctx, uuid, pubKey, 0).Err()
	if err != nil {
		uuidWithHyphen := uuid.New()
		uuid = strings.Replace(uuidWithHyphen.String(), "-", "", -1)
		err := client.Set(ctx, uuid, pubKey, 0).Err()
	}

	argLength := len(os.Args[1:])
	if argLength < 4 {
		fmt.Printf("Not enough arguments. Need platform, architecture, callback URL, and output file name \n")
	}
	mydir, _ := os.Getwd()
	var (
		err       error
		fp        *os.File
		templates *template.Template
	)

	values := appValues{}
	fmt.Printf("Dagger builder \n")

	values.CallBack = os.Args[4]
	values.AppName = os.Args[3]

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

	fmt.Printf(" Done  building template \n")

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

	fmt.Printf(" Done compiling \n ")

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
