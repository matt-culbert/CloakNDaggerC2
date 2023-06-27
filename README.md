# Cloak & Dagger

This is an evolution of the original Switchblade C2. Cloak refers to the C2 backend, hiding behind an mTLS reverse proxy, and Dagger is the implant which utilizes syscalls for command execution.

Run redis in a Docker container with ```docker run --name redis -p 6379:6379 -d redis```

When you run a command, you need to specify the UUID of the implant every time. To get a list of UUIDs in the redis db, enter '''5'''

![example](/img/guide/example.png)

Current commands:
- ```pwd``` gets the current working directory
- ```gcu``` gets the current user
- ```rc``` runs a command through the terminal, this can be anything (Still working on making commands work that are more than one word. So '''whoami''' works fine but '''cat /etc/passwd''' has issues
- ```rd``` reads the supplied directory. Use it with '''rd <directory path>'''

To generate keys, run the ```crypto.py``` app and copy the public key PEM contents into the implant.go file


Todo: 

- [ ] Generate shellcode from the controller by adding ```go build -buildmode=pie -o shellcode.bin .\beacon.go```
- [ ] Add a way to change the config of the beacon
- [ ] Add RSA key gen to builder. This will be needed for the reverse proxy as well
- [ ] XOR enc IP
- [ ] Clean up the usability from the console
- [ ] RSA encrypt results of command execution with public key
- [ ] XOR enc incoming commands with a pre shared secret, possibly the public key??
