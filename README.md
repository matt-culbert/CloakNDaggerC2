# Before first use
This covers the initial setup of required things like the Python/Go requirements and how to generate a user.

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
cd ./Server; python pw_hash.py
```
Generate the SSL cert for the server to secure connections with
```bash
cd ./Server; openssl req -new -x509 -keyout server.pem -out server.pem -days 365 -nodes
```

### Run the application
1) First, start the server
```bash
cd ./Server; python server.py
```
2) After the server is started, start the client and enter the username/password you generated. The server needs to be run first since the client tries to authenticate after you enter your details.
```bash
python Client/client.py
```

# Compiling the implant
You should use the Makefile when compiling an implant. It has several requirements and these are laid out below.

### Required build arguments
There are tags and ldflags that setup things like implant ID and enable supported features. Tags are also used to define what communication method to use. The implant relies on a configuration file for the rest of its settings. If compression is enabled, then this configuration file is saved as a bin. Otherwise, it is saved as a simple JSON struct.

#### Compile flags
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
```
The next set of flags are required for determining which communication method to use
```bash
# Use HTTP for communication
withHttp
# Use DNS for communication
withDns
```

When not using the makefile, this looks like the following:
```bash
go build -ldflags <ldflags> -tags <features> ./base_config/daemon
```
OR
```bash
go build -trimpath -ldflags "-X main.CompUUID=1234 -s -w" -tags "withComp withDns" ./base_config/daemon
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
