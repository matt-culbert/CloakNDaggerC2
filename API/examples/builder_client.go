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
	ADDRESS = "localhost:50053" // builder listener port
)

// This will be used to regist the implant with the DB frontend
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
	conn, err := grpc.Dial(ADDRESS, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect : %v", err)
	}
	fmt.Printf("0")

	defer conn.Close()
	fmt.Printf("1")

	c := pb.NewBuilderClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()
	res, err := c.StartBuilding(ctx, &pb.BuildRoutine{Platform: "windows", Architecture: "amd64", Name: "string_hashing", ListenerAddress: "https://test.culbertreport:8000"})

	if err != nil {
		log.Fatalf("could not save implant: %v", err)
	}
	fmt.Println(res.GetCode())

}
