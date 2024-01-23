//go:build !ignorevet
// +build !ignorevet

package main

// This is a testing example demonstrating a set command to set implant details in the API

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "CloakNDaggerC2/dagger/proto/daggerProto"

	"google.golang.org/grpc"
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
	conn, err := grpc.Dial("localhost:50055", grpc.WithInsecure(), grpc.WithBlock())
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
