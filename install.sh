go mod tidy

go build -v

docker run --name redis -p 6379:6379 -d redis
