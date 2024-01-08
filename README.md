# Cloak & Dagger

This is currently in a broken state as I push over to an API based infrastructure. If you would like a version that worked in the past, I believe this version should be ok https://github.com/matt-culbert/CloakNDaggerC2/commit/6a775b2ac07aa60659d8810cfc6389805430e7cf

![logo](/img/guide/cnd8.jpg)

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
- ```groups``` returns the SID of all local groups the user is in
- ```pid``` returns the current process ID

### Known issues:
On Kali, change the redis host in the controller and listener to 127.0.0.1 from localhost.

### Todo: 

Near term goals for long term viability
- [ ] Tidy up code body for long term survivability
  - [ ] This should include profile support upon first launch of the controller which:
          - Allows defining the listener address
          - Round robin listening addresses
          - Functions in implant should be broken out to sub folders and utilize the package tag for compilation - I think this is the next goal

Core items:

These are the items that need to be done to make the framework actually usable
- [ ] Add absolute paths when calling applets from other folders....
- [ ] Adjust so that failure to reach C2 or bad response doesn't crash implant
- [ ] Obfuscation of function names at generation time. This way each sample has unique hashes
- [ ] Periscope had a great idea - canary URLs. If a canary URL is queried then the redirector just returns 404 to the investigator. These will only be seen people decompiling looking for strings so lets just add their info to a block list that lasts as long as the campaign.
- [ ] Add a historical context for report exporting from Redis for all commands run on a target.
- [ ] Take unique system ID in dropper and change the builder to generate an EXE linked solely to this ID.
- [ ] Add profile support for different URL paths. Then listeners can just pull from here each time they're started and implants will pull from here on generation
- [ ] Add multiple call back methods
- [ ] Token theft

Future tasks:

These are tasks to be completed that will make the framework appealing to the target audience
- [ ] Add a .NET appdomain function for running tools like SeatBelt. We can load these recieved binaries into a byte array. But this approach only allows one loaded at a time
- [ ] Add persistence for the binaries sent over, so they only need to be sent once. Maybe encrypt with a key then decrypt to run, then encrypt with the current time stamp as a new key? who knows
- [ ] Build in lateral movement functionality. So LSASS dumping and SMB
- [ ] General process memory dumping ability
- [ ] Construct a persistent directory tree generated from interacting with implants over time
