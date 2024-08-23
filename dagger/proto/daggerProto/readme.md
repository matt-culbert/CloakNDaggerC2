Troubleshooting errors
---
If you encounter errors when rebuilding the proto file, some common issues include:

Missing the following packages
- go get -u google.golang.org/protobuf
- go get -u google.golang.org/grpc

Make sure protoc is in your environment path

After modifying the proto file, remember to rebuild it
---
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative dagger.proto