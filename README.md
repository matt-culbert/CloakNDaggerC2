# Cloak & Dagger

![logo](/img/guide/cnd8.jpg)

There are keys included here, they're purely for testing. You should expect these to be burned and thus generate your own.

If you're gonna skip running the install script to set everything up, you're gonna have a bad time. 

## So what is this?

Simply put, CloakNDagger is a framework designed around the use of public/private RSA key pairs to sign and authenticate commands being executed. This prevents MiTM interception of calls and ensures opsec during delicate operations. Any command sent to the implant to be executed must be signed and that signature must be verified before execution. The implant also uses fingerprinting of the listeners TLS certs in order to verify that they are indeed correct when every request for a command is sent. This is intended to be a redundant failure point, if one these checks stops working correctly you still have the other that you can rely on to verify authenticity.

[A quote demonstrating exactly what I want to solve](https://assume-breach.medium.com/im-not-a-pentester-and-you-might-not-want-to-be-one-either-8b5701808dfc)

> Here’s another thing, got a dope implant? Think you’re going to drop EXEs on a target? Think again. I wasn’t allowed to drop anything to disk when I was a pentester. Why? Because I “might” forget about them ...

With CloakNDagger you can leave those implants running for the rest of time and, until someone breaks RSA, only you can send executable commands to them.

Commands are primarily run through the os and os/user packages from Go. These allow you to perform many operations without needing to go through the command interpreter. This is because they have abstracted SYSCALLs away from the user and do the heavy lifting of implementing them for you. 

## Requirements

Go 1.20 +

Docker

The certs and PEM files are required, but like I said they should be considered burned. The global files are what are used for signing commands and authenticating the signature, then the server cert files are what are used for serving the TLS connection and verifying the fingerprint. 

## Use

Run redis in a Docker container with ```docker run --name redis -p 6379:6379 -d redis```

When you run the install script on first use, this is started alongside it. But for the future starts, you'll need to make sure Docker is running redis.

Once the script builds the main program, run it through ```./CloakNDaggerC2``` and voila everything starts up!

## Known issues:
If you look at the compiled implant in a debugger and search for http strings, you'll quickly find the listener address. This is because there is a non failing error to do with an incorrect header. Trying to fix that but for now it's a great point to analysts to look at and find C2's.

Upon building your first implant for a platform, you will get an error on the status and control will return to the main function. Then after a moment the UUID will be displayed and a message that it was added to the DB.

The fingerprint is hashed on the implant side using a string hashing method that is not second preimage resistant or collision resistant. This could lead to failure to properly verify down the line if someone can generate a hash of another message that equals this hash (H(x1) == H(x2)) <- I'm unsure if this will be addressed or not, I need to do some big math on the likelihood and impact versus the gains from string hashing.

If you try to create a listener, get to the URL handler section, and exit, it will still try to serve on that port causing issues when you attempt to start another listener. 

Implants exiting when the C2 is not available has cropped again, looking to smush this bug.
