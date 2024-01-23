//go:build !ignorevet
// +build !ignorevet

package main

// This is a testing example of how you can use GRPC to write your own client to interact with the builder

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "CloakNDaggerC2/dagger/proto/daggerProto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect : %v", err)
	}

	defer conn.Close()

	c := pb.NewBuilderClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()
	res, err := c.StartBuilding(ctx, &pb.BuildRoutine{Platform: "windows", Architecture: "amd64", Name: "string_hashing", ListenerAddress: "https://test.culbertreport:8000"})

	if err != nil {
		log.Fatalf("could not save implant: %v", err)
	}
	fmt.Println(res.GetCode())

}
