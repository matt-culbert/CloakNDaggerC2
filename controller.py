import listener
import redis

conn = redis.StrictRedis(host='localhost', port=6379, db=0)

def updateCommand(uuid, command):
    conn.hset('UUID', uuid, command)

def searchUUID(uuid):
    conn.hget('UUID', uuid)

while True:
    inp = input('(1)Enter command / (2)List all')
    if inp == '1':
        uuid = input('UUID: ')
        comm = input('Command: ')
        updateCommand(uuid, comm)
    else:
        #print(conn.keys()) # UUID is the key but we want values from the key
        print(conn.hgetall('UUID')) # We're searching by hash values here
