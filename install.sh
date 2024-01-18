go mod tidy

go_src_dir="src"

go build "./$go_src_dir/Listeners/controller.go Listeners/listener_func.go"
go build "./$go_src_dir/API/API_Listener.go"
go build "./$go_src_dir/Builder/builder.go"

docker run --name redis -p 6379:6379 -d redis
