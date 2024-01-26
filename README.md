# Cloak & Dagger

![logo](/img/guide/cnd8.jpg)

A C2 framework designed around the use of public/private RSA key pairs to sign and authenticate commands being executed. This prevents MiTM interception of calls and ensures opsec during delicate operations.

This is an evolution of the original Switchblade C2. Cloak refers to the C2 backend and Dagger is the implant.

There are keys included here, they're purely for testing. You should expect these to be burned and thus generate your own.

If you're gonna skip running the install script to set everything up, you're gonna have a bad time. 

### Requirements

Go 1.20 +

Docker

### Use

Run redis in a Docker container with ```docker run --name redis -p 6379:6379 -d redis```

When you run the install script on first use, this is started alongside it. But for the future starts, you'll need to make sure Docker is running redis.

Once the script builds the main program, run it through ```./CloakNDaggerC2``` and voila everything starts up!

### Known issues:
If you look at it in a debugger and search for http strings, you'll quickly find the listener address. This is because there is a non failing error to do with an incorrect header. Trying to fix that but for now it's a great point to analysts to look at and find C2's.

Upon building your first implant for a platform, you will get an error on the status and control will return to the main function. Then after a moment the UUID will be displayed and a message that it was added to the DB.


