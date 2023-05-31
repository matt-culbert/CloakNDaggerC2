# Cloak & Dagger

This is an evolution of the original Switchblade C2. Cloak refers to the C2 backend, hiding behind an mTLS reverse proxy, and Dagger is the implant which utilizes syscalls for command execution.

Run redis in a Docker container with ```docker run --name redis -p 6379:6379 -d redis```

Todo: 

- [ ] Generate shellcode from the controller by adding ```go build -buildmode=pie -o shellcode.bin .\beacon.go```
- [ ] Add a way to change the config of the beacon
- [ ] Add RSA key gen to builder. This will be needed for the reverse proxy as well
