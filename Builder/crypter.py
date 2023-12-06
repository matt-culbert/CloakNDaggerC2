from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives import serialization
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
  with open("../global.pem", "wb") as key_file:
    key_file.write(pem)
  with open("../global.pem", "rb") as key_file:
    private_key = serialization.load_pem_private_key(key_file.read(), password=None)

  with open("../global.pub.pem", "wb") as public_file:
    public_file.write(pem_public_key)

except Exception as e:
  print(e)


