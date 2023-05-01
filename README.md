# SwitchBladev3

3rd times the charm right?

Insetad of using HTML files, we're just doing everything in redis.

Run redis in a Docker container with ```docker run --name redis -p 6379:6379 -d redis```

Todo: 

[] Generate shellcode from the controller by adding ```go build -buildmode=pie -o shellcode.bin .\beacon.go```

[] Add a way to change the config of the beacon
