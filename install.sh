go mod tidy

go build -o "./Listeners/" "./Listeners/controller.go" "Listeners/listener_func.go"
go build -o "./API/" "./API/API_Listener.go"
go build -o "./Builder/" "./Builder/builder.go"

docker run --name redis -p 6379:6379 -d redis
