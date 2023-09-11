from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives import serialization
import redis
import datetime
import sys
import json
size=512

try:
  private_key = rsa.generate_private_key(
    public_exponent=65537,
    key_size=2048,
  )
  pem_public_key = private_key.public_key().public_bytes(
    encoding=serialization.Encoding.PEM,
    format=serialization.PublicFormat.SubjectPublicKeyInfo
  )
  pem = private_key.private_bytes(
    encoding=serialization.Encoding.PEM,
    format=serialization.PrivateFormat.TraditionalOpenSSL,
    encryption_algorithm=serialization.NoEncryption()
  )
  with open("keys/"+sys.argv[1]+".pem", "wb") as key_file:
    key_file.write(pem)
  with open("keys/"+sys.argv[1]+".pem", "rb") as key_file:
    private_key = serialization.load_pem_private_key(key_file.read(), password=None)

  with open("keys/"+sys.argv[1]+".pub.pem", "wb") as public_file:
    public_file.write(pem_public_key)

  conn = redis.StrictRedis(host='localhost', port=6379, db=0)

  structure = {
      "WhoAmI": f"{sys.argv[1]}",
      "Nonce": f"0",
      "Signature": "0",
      "Retrieved": "0",
      "Command": "0",
      "LastInteraction": "0",
      "LastCheckIn": "",
      "Result": "0",
      "GotIt": "0"
  }
  structure = json.dumps(structure)  # Dump the json
  # Write the message value to the beacon:UUID key
  conn.hset('UUID', sys.argv[1], structure)

  print("setup UUID")

except Exception as e:
  print(e)


