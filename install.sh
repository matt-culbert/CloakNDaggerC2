go mod tidy

go build Listeners/controller.go Listeners/listener_func.go
go build API/API_Listener.go
go build Builder/builder.go

docker run --name redis -p 6379:6379 -d redis