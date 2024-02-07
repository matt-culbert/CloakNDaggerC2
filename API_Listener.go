// package main is required for any executable, but other sub pieces don't need to be main
package main

import (
	// go mod init dagger/proto/daggerProto - which led to the duplicate import path
	"context"
	"encoding/json"
	"fmt"

	pb "CloakNDaggerC2/dagger/proto/daggerProto"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

// This program will be running and recieving the data marshalled by the proto file
// We want to unmarshal it then load it into Redis

// This works all well and good for adding test data
// The next goal will be to :
// 1) Setup a listening gRPC server
// 2) Write the calls from there to the Redis DB
// 3) Do this with SET and HSET - SET will be for historical context

// The reason for this project is to abstract the complicated pieces of committing this to the DB away from the user
// Now, applications just need to make a gRPC call to this service
// Users can also query this service with whatever they want, adding the option for automation through Python scripts that can update the fields and read values

// It's looking like we're going to need to move the Redis functions and associated data out of main
// In main() we'll call the server, serve up the listener
// In the listener function we'll ingest the sent data and return a status code
// protoc --go-grpc_out=. dagger.proto
// The above allowed us to generate go specific gRPC code, which included the service

// The server runs and listens for incoming connections,
// Now to write the test client to send the formatted data to be inserted
// Then once that's tested and works good, we need to update the controller to use this
// And the builder needs to be pointed to this

type RecieveImpUpdate struct {
	pb.UnimplementedUpdateRecordServer
}

type hgetUUID struct {
	pb.UnimplementedHgetRecordServer
}

type Rkey struct {
	pb.UnimplementedRemoveServer
}

type GetAll struct {
	pb.UnimplementedGetAllServer
}

type ImplantLayout struct {
	UUID        string   `redis:"UUID"`
	Whoami      string   `redis:"WhoAmI"`
	Signature   string   `redis:"Signature"`
	Retrieved   int32    `redis:"Retrieved"`
	Command     string   `redis:"Command"`
	LastCheckIn string   `redis:"CheckIn"`
	Result      string   `redis:"Result"`
	GotIt       int32    `redis:"GotIt"`
	Ignored     struct{} `redis:"-"`
}

// This struct is a copy of the ImplantStruct but with redis field tags
type IIDScan struct {
	UUID        byte     `redis:"UUID"`
	Whoami      string   `redis:"WhoAmI"`
	Signature   string   `redis:"Signature"`
	Retrieved   int32    `redis:"Retrieved"`
	Command     string   `redis:"Command"`
	LastCheckIn string   `redis:"CheckIn"`
	Result      string   `redis:"Result"`
	GotIt       int32    `redis:"GotIt"`
	Ignored     struct{} `redis:"-"`
}

func (s *Rkey) RemKey(ctx context.Context, in *pb.DelKey) (*pb.ResponseCode, error) {
	fmt.Println("Removing key")
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	err := client.Del(ctx, in.Key)
	if err != nil {
		//fmt.Println("Error on removing uuid")
		ResponseCode := &pb.ResponseCode{
			Code: 1,
		}
		return ResponseCode, nil
	} else {
		ResponseCode := &pb.ResponseCode{
			Code: 0,
		}
		// We return the response code of the action to update the redisdb
		// If success, we return a 0. If the operation to HSet fails, we return a 1
		// Could possibly use a more verbose error message in the future
		// But for now we know the error occurs when HSet'ng so that's good enough
		return ResponseCode, nil
	}
}

func (s *GetAll) GetAll(ctx context.Context, in *pb.GetKey) (*pb.DbContents, error) {
	// This will allow us to search the DB for every key that matches "UUID" and dump their values
	// This will also allow us to seach the DB for every subfield in that entry and return the LastCheckIn time
	// We need to return the array of keys and the array of LCI
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	var entries []string

	iter := client.HScan(ctx, in.GetKey(), 0, "*", 0).Iterator()

	if iter.Err() != nil {
		return nil, iter.Err()
	}

	for iter.Next(ctx) {
		// Once we have a key we need to then do a query for that key, which will allow us to get individual elements from it
		var scanModel IIDScan
		// Stuck right now on unmarshalling into the struct....
		// The Result() method gets the values from the HGet call
		vals, _ := client.HGet(ctx, "UUID", iter.Val()).Result()
		// We need to move vals into a structured form now
		// So this returns an error but still works for what's needed....
		// I'm a dum dum
		_ = json.Unmarshal([]byte(vals), &scanModel)

		// We should now have the data mapped to the scan model
		res := scanModel.LastCheckIn
		entry := res + ", " + iter.Val() + "; "
		entries = append(entries, entry)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	// We return the array of entries which should be UUID, LCI
	return &pb.DbContents{Res: entries}, nil
}

// The name 'SendUpdate' is important here as that's the function we defined in the UpdateRecord service
func (s *RecieveImpUpdate) SendUpdate(ctx context.Context, in *pb.UpdateObject) (*pb.ResponseCode, error) {
	// The Redis connection string
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// This below struct holds the info that we're getting
	ImplantInfo := &pb.UpdateObject{
		UUID:        in.GetUUID(),
		Whoami:      in.GetWhoami(),
		Signature:   in.GetSignature(),
		Retrieved:   in.GetRetrieved(),
		Command:     in.GetCommand(),
		LastCheckIn: in.GetLastCheckIn(),
		Result:      in.GetResult(),
		GotIt:       in.GetGotIt(),
	}
	// Here we marshal ImplantInfo into a format, "data" gets the marshaled data written to it in the format of the ImplantID struct
	data, _ := proto.Marshal(ImplantInfo)

	// So we have the data marshalled into a format we can read, we now need to unmarshal it then write it to the Redis DB
	// This creates a new struct from which we can unmarshal data to
	unmarshaled_data := &pb.UpdateObject{}
	// Deserialize it
	proto.Unmarshal(data, unmarshaled_data)

	// This is something that needs solving
	// Why are we marshalling, unmarshalling, then marshalling into json...

	// This is overkill for now and needs a better process in the future, but
	// Lets now map each element of the unmarshaled_data to the ImplantData struct
	ImpData, _ := json.Marshal(ImplantLayout{UUID: unmarshaled_data.UUID, Whoami: unmarshaled_data.Whoami, Signature: unmarshaled_data.Signature,
		Retrieved: unmarshaled_data.Retrieved, Command: unmarshaled_data.Command, LastCheckIn: unmarshaled_data.LastCheckIn, Result: unmarshaled_data.Result,
		GotIt: unmarshaled_data.GotIt})

	// Write the deserialized data
	// hset (ctx, hash key, field name, data)
	// We need to turn the unmarshaled data into a struct that can be used here
	// The issue right now is that this can't be saved as is, we need to marshal it to something
	// Online suggestiosn are to marshal to JSON which we do on the Python side
	// So let's skip the saving as unmarshaled data and save the marshaled stuff
	// Erm if we keep it marshaled as proto, this is always a byte array which we have trouble unmarshalling
	// I see now, in Python I set it so the fields are flipped, UUID is the hash and the unmarshaled UUID is the key
	//_ = client.Set(ctx, unmarshaled_data.UUID, ImpData, 0).Err()
	err := client.HSet(ctx, "UUID", unmarshaled_data.UUID, ImpData).Err()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error on HSet for client")
		ResponseCode := &pb.ResponseCode{
			Code: 1,
		}
		return ResponseCode, nil
	} else {
		ResponseCode := &pb.ResponseCode{
			Code: 0,
		}
		// We return the response code of the action to update the redisdb
		// If success, we return a 0. If the operation to HSet fails, we return a 1
		// Could possibly use a more verbose error message in the future
		// But for now we know the error occurs when HSet'ng so that's good enough
		return ResponseCode, nil
	}

}

func (s *hgetUUID) Hget(ctx context.Context, in *pb.GetUUID) (*pb.UpdateObject, error) {
	// This function performs an hget for the UUID passed to it
	// It returns the correspdong struct for the implant info
	// general get will be harder because it will return an undefined number of rows....
	// The Redis connection string
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// UUID, searching for tags command, checkin, and result
	// unmarshaled_data.UUID represents the hash key we are scanning for, then getting the fields at the tags
	var scanModel IIDScan
	// The Result() method gets the values from the HMGet call
	vals, err := client.HGet(ctx, "UUID", in.GetUUID()).Result()
	// We need to move vals into a structured form now
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal([]byte(vals), &scanModel)

	// We should now have the data mapped to the scan model

	res := &pb.UpdateObject{
		UUID: in.GetUUID(), Whoami: scanModel.Whoami, Signature: scanModel.Signature,
		Retrieved: 0, Command: scanModel.Command, LastCheckIn: scanModel.LastCheckIn, Result: scanModel.Result,
		GotIt: 0,
	}
	// any errors should have been caught before this section during the hget function
	// so I'm confident that we can return the struct without checking for an error here
	return res, nil
}
