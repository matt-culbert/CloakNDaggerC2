### Before first use
This covers the initial setup of required things like the Python requirements and how to generate a user.

Install required Python modules
```bash
pip install -r requirements.txt
```
Install required Go packages
```bash
cd ./Implant; go mod tidy
```
Generate a user by running the pw_hash.py script
```bash
python Server/pw_hash.py
```
Generate the SSL cert for the server to secure connections with
```bash
openssl req -new -x509 -keyout server.pem -out server.pem -days 365 -nodes
```

### Run the application
1) First, start the server
```bash
python Server/server.py
```
2) After the server is started, start the client and enter the username/password you generated. The server needs to be run first since the client tries to authenticate after you enter your details.
```bash
python Client/client.py
```

### Powershell commands for simulating implant
These have proven handy for emulating implant testing to the server without having to run an implant itself. 
```powershell
$postParams = @{"msg" = "7465:7374::"}
$jsonPost = $postParams | ConvertTo-Json 
$headers = New-Object "System.Collections.Generic.Dictionary[[String],[String]]"
$headers.Add('Content-type','Application/Json')
Invoke-WebRequest -Uri http://127.0.0.1:5000/1234 -Method GET
Invoke-WebRequest -Uri http://127.0.0.1:5000/1234 -Method POST -Body $jsonPost -Headers $headers
```

### Compiling Go exe
You should read/use the Makefile.
#### Makefile CLI arguments
The make file takes arguments from the CLI under certain scenarios. When you're not building the default build, you need to pass in the METHOD, which tells make what communication mode to compile for.
```bash
make withLua METHOD=withHttp # Example to compile for HTTP comms
```
#### Compiling without the Makefile
If you're curious about what's supported or you want to compile the implant manually, this covers it.
There are tags and ldflags that setup things like the callback URLs, implant ID, and enable supported features.
#### Compile flag options
Most options are read from the implant config file. The one requiring input is the UUID to compile with.
```bash
# Implant UUID
# Expects a random but unique 4 digit integer
-X main.CompUUID 
```
#### Compile tag options
```bash
# Enable zlib compression
withComp 
# Enable support for Lua scripting
withLua 
# ----------------------
# The next set of flags are required, use one of them
# Use HTTP for communication
withHttp
# Use DNS for communication
withDns
```
#### Syntax
```bash
go build -ldflags <ldflags> -tags <features> ./Implant/daemon
```
#### Example
```bash
go build -trimpath -ldflags "-X main.CompUUID=1234 -s -w" -tags "withComp withDns" ./preprocessor/daemon
```

### User customization
Theres several profile options and listener options that can be configured.
#### Pre-shared encryption keys
- PSK1 is used for authenticating the HMACs sent by the server
- PSK2 is used by the implant for generating authentication tokens
  - These tokens are required for listeners set to require authentication before returning queued commands
#### Custom listeners
Users can define their own listeners by building a new Flask Blueprint in the Server/blueprints folder. An example is provided there for reference.
After the new Blueprint is saved, add the route to the s_conf.json file and refresh the listeners.
