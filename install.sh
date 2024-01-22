go mod tidy

go build controller.go listner_func.go API_Listener.go builder.go

docker run --name redis -p 6379:6379 -d redis
