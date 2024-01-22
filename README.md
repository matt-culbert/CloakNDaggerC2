# Cloak & Dagger

![logo](/img/guide/cnd8.jpg)

A C2 framework designed around the use of public/private RSA key pairs to sign and authenticate commands being executed. This prevents MiTM interception of calls and ensures opsec during delicate operations.

This is an evolution of the original Switchblade C2. Cloak refers to the C2 backend and Dagger is the implant.

There are keys included here, they're purely for testing. You should expect these to be burned and thus generate your own.

### Setup

Run redis in a Docker container with ```docker run --name redis -p 6379:6379 -d redis```

This should be started automatically by the install script, but for future use you will need this.

### Generating implants

The menu option for generating implants is ```1``` so hit that and follow the onscreen instructions.

### Interacting with implants

When you run a command, you need to specify the UUID of the implant every time. 

Current commands:
- ```pwd``` gets the current working directory
- ```gcu``` gets the current user
- ```rc``` runs a command through the terminal, this can be anything 
- ```rd``` reads the supplied directory. Use it with '''rd <directory path>'''
- ```terminal``` allows you to run terminal commands - NOT OPSEC SAFE
- ```groups``` returns the SID of all local groups the user is in
- ```pid``` returns the current process ID

### Known issues:
On Kali, change the redis host in the controller and listener to 127.0.0.1 from localhost.


