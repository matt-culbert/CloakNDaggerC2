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

	pb "dagger/proto/daggerProto/dagger/proto/daggerProto"

	"google.golang.org/grpc"
)

const (
	ADDRESS = "localhost:50055"
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
	conn, err := grpc.Dial(ADDRESS, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect : %v", err)
	}

	defer conn.Close()

	c := pb.NewUpdateRecordClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	data := []ImplantInfo{
		{UUID: "test1", Whoami: "1", Signature: "1", Retrieved: 1, Command: "1", LastCheckIn: "1", Result: "1", GotIt: 0},
	}

	for _, info := range data {
		res, err := c.SendUpdate(ctx, &pb.UpdateObject{UUID: info.UUID, Whoami: info.Whoami, Signature: info.Signature,
			Retrieved: info.Retrieved, Command: info.Command, LastCheckIn: info.LastCheckIn, Result: info.Result,
			GotIt: info.GotIt})

		if err != nil {
			log.Fatalf("could not save implant: %v", err)
		}
		fmt.Println(res.GetCode())
	}
}
