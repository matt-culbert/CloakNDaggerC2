from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives import hashes
import sys
import binascii
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
  with open("keys/test/test_priv.pem", "wb") as key_file:
    key_file.write(pem)
  with open("keys/test/test_priv.pem", "rb") as key_file:
    private_key = serialization.load_pem_private_key(key_file.read(), password=None)

  pub = private_key.public_key()

  with open("keys/test/test_pub.pem", "wb") as public_file:
    public_file.write(pub)

except:
  pass

