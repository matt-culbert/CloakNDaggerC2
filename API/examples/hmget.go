package main

// This will be the gRPC client writing to the API_Listener.go
// The end goal will be that the Python controller and other apps can write to it
// But this is for testing. It won't be needed for long

import (
	// go mod init dagger/proto/daggerProto - which led to the duplicate import
	"context"
	"fmt"
	"log"
	"time"

	pb "CloakNDaggerC2/dagger/proto/daggerProto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ADDRESS = "localhost:50055" // hmget listener port
)

type ImplantInfo struct {
	UUID        string
	Whoami      string
	Signature   string
	Retrieved   int32
	Command     string
	LastCheckIn string
	Result      string
	GotIt       int32
}

func main() {
	// currently freezing on this section
	// ugh it's freezing because the server is serving the first API method first not ours
	conn, err := grpc.Dial(ADDRESS, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect : %v", err)
	}

	defer conn.Close()

	c := pb.NewHgetRecordClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	UUID := "test1"

	res, err := c.Hget(ctx, &pb.GetUUID{UUID: UUID})

	if err != nil {
		log.Fatalf("could not find UUID: %v", err)
	}
	fmt.Printf("The last command was %s \n", res.Command)
	fmt.Printf("The result of whoami was %s \n", res.Whoami)

}
