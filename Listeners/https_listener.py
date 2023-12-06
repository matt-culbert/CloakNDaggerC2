import sys
from datetime import datetime
from flask import *
import json
import redis
import string
from dataclasses import dataclass
#import API.redis_calls

@dataclass
class dbParameters:
    WhoAmI: str
    Nonce: str
    Signature: str
    Retrieved: str  # Reset retrieved so we know the command was picked up
    Command: str
    LastInteraction: str
    LastCheckIn: str
    Result: str
    GotIt: str


conn = redis.StrictRedis(host='localhost', port=6379, db=0)
app = Flask(__name__)


def updateDB(data_struct, uuid):
    structure = {
        "WhoAmI": data_struct.WhoAmI,
        "Nonce": data_struct.Nonce,
        "Signature": data_struct.Signature,
        "Retrieved": data_struct.Retrieved,
        "Command": data_struct.Command,
        "LastInteraction": data_struct.LastInteraction,
        "LastCheckIn": datetime.today().strftime('%Y-%m-%d %H:%M:%S'),
        "Result": data_struct.Result,
        "GotIt": data_struct.GotIt
    }
    structure = json.dumps(structure)  # Dump the json
    # Write the message value to the beacon:UUID key
    conn.hset('UUID', uuid, structure)


@app.route('/', methods=['GET'])
# We no longer need this function, registration happens in the generator
def home():
    print('start')
    # This handles initial registration
    # The beacon receives the initial reg and adds it to the db
    if request.method == 'GET':
        uuid = request.headers['APPSESSIONID']
        whoami = request.headers['Res']
        dataSet = dbParameters(whoami, 0, 0, 0, 0, 0, datetime.today().strftime('%Y-%m-%d %H:%M:%S'), 0, 0)
        if set(uuid).difference(string.ascii_letters + string.digits):
            # We're not going to bother with input sanitization here
            # If we receive special characters just drop it entirely
            pass
        else:
            updateDB(dataSet, uuid)
            return ''


@app.route('/session', methods=['GET'])
def session():
    print('session')
    # This function handles the beacon requesting a command
    if request.method == 'GET':
        uuid = request.headers['APPSESSIONID']
        if set(uuid).difference(string.ascii_letters + string.digits):
            # We're not going to bother with input sanitization here
            # If we receive special characters just drop it entirely
            pass
        else:
            connt = conn.hget('UUID', uuid)  # Get the struct
            connt = connt.decode()  # Decode it from bytes
            connt = json.loads(connt)  # it's returned as string so convert it to dict
            structure = json.dumps(connt)  # Dump the dict to json
            connector = json.loads(structure)  # Load it into a new var
            command = connector["Command"]  # Grab the command var from the object
            LastInteraction = datetime.today().strftime('%Y-%m-%d')
            lastCheckIn = connector["LastCheckIn"]
            result = connector["Result"]
            whoami = connector["WhoAmI"]
            nonce = connector["Nonce"]
            signature = connector["Signature"]
            print(signature)
            #whoami, nonce, signature, retrieved, command, last interaction, last check in, result, got it
            dataSet = dbParameters(whoami, nonce, signature, "1", "0", LastInteraction,
                                   lastCheckIn, result, "0")
            if set(uuid).difference(string.ascii_letters + string.digits):
                # We're not going to bother with input sanitization here
                # If we receive special characters just drop it entirely
                pass
            else:
                updateDB(dataSet, uuid)

            resp = Response(
                response=command, status=302, mimetype="text/plain")
            resp.headers['Verifier'] = signature
            resp.headers['Nonce'] = nonce
            print(signature)
            return resp  # we've really got to RC4 encrypt this


@app.route("/schema", methods=['GET'])
def schema():
    # This function handles beacons returning data
    # Grab the appsessionid value from the headers
    # We need to grab the corresponding private key from the key store to decrypt incoming messages with
    # Set a default value on checkin to create the hostname/whoami to identify the beacon [needs testing]
    uuid = request.headers['APPSESSIONID']
    result = request.headers['Res']
    #nonce = request.headers['nonce']

    command = conn.hget('UUID', uuid)  # Get the struct
    command = command.decode()  # Decode it from bytes
    command = json.loads(command)  # it's returned as string so convert it to dict
    structure = json.dumps(command)  # Dump the dict to json
    connector = json.loads(structure)  # Load it into a new var
    LastInteraction = connector["LastInteraction"]
    whoami = connector["WhoAmI"]
    # whoami, nonce, signature, retrieved, command, last interaction, last check in, result, got it
    dataSet = dbParameters(whoami, "0", "0", "1", "0", LastInteraction,
                           datetime.today().strftime('%Y-%m-%d %H:%M:%S'), result, "1")
    if set(uuid).difference(string.ascii_letters + string.digits):
        # We're not going to bother with input sanitization here
        # If we receive special characters just drop it entirely
        pass
    else:
        updateDB(dataSet, uuid)
        return ''


def serve():
    context = ('testServer.crt', 'testServer.key')
    app.run('test.culbertreport', 8000, ssl_context=context)


if __name__ == "__main__":
    serve()