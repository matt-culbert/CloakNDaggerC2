# Cloak & Dagger

![logo](/img/guide/cnd8.png)

A C2 framework designed around the use of public/private RSA key pairs to sign and authenticate commands being executed. This prevents MiTM interception of calls and ensures opsec during delicate operations.

This is an evolution of the original Switchblade C2. Cloak refers to the C2 backend and Dagger is the implant.

### Setup

Run redis in a Docker container with ```docker run --name redis -p 6379:6379 -d redis```

### Generating implants

Generate an implant by building ```builder.go``` or just by running it with go.

It takes 4 commands; platform, architecture, output file name, and callback URL. It will handle calling ```crypter.py``` with the UUID generated to make the RSA keys. These are then stored in the keys directory and referenced by the controller.

### Interacting with implants

When you run a command, you need to specify the UUID of the implant every time. To get a list of UUIDs in the redis db, enter '''5'''

![example](/img/guide/example.png)

Current commands:
- ```pwd``` gets the current working directory
- ```gcu``` gets the current user
- ```rc``` runs a command through the terminal, this can be anything (Still working on making commands work that are more than one word. So '''whoami''' works fine but '''cat /etc/passwd''' has issues
- ```rd``` reads the supplied directory. Use it with '''rd <directory path>'''
- ```terminal``` allows you to run terminal commands - NOT OPSEC SAFE

### Known issues:
On Kali, change the redis host in the controller and listener to 127.0.0.1 from localhost.

### Todo: 

Core items:

These are the items that need to be done to make the framework actually usable
- [ ] Clean up the CLI output
- [ ] Periscope had a great idea - canary URLs. If a canary URL is queried then the redirector just returns 404 to the investigator. These will only be seen people decompiling looking for strings so lets just add their info to a block list that lasts as long as the campaign.
- [ ] Add a historical context for report exporting from Redis for all commands run on a target. Probably with MongoDB so that we can also encrypt it?
- [x] Make an install script
- [x] Registration should occur when the implant is compiled, the listener can then check the redis DB for the corresponding private key. Can also do this on the edge with an nginx reverse proxy
- [ ] Add profile support for different URL paths. Then listeners can just pull from here each time they're started and implants will pull from here on generation
- [x] Change the listeners to accept arbitrary URLs for callback
- [ ] I really need to break out the API calls into a seperate file. It's large and clunky and looking at it makes me sad :(
- [ ] List PID and PPID
- [ ] Token theft

Future tasks:

These are tasks to be completed that will make the framework appealing to the target audience
- [ ] Add a .NET appdomain function for running tools like SeatBelt. We can load these recieved binaries into a byte array. But this approach only allows one loaded at a time
- [ ] Add persistence for the binaries sent over, so they only need to be sent once. Maybe encrypt with a key then decrypt to run, then encrypt with the current time stamp as a new key? who knows
- [ ] Build in lateral movement functionality. So LSASS dumping and SMB
- [ ] General process memory dumping ability
- [ ] Construct a persistent directory tree generated from interacting with implants over time
