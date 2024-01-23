//go:build !ignorevet
// +build !ignorevet

package main

// This is a testing example demonstrating getting a single result

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "CloakNDaggerC2/dagger/proto/daggerProto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
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
