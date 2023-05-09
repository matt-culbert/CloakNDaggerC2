from Crypto.PublicKey import RSA
import base64

secret_code = "Unguessable"
key = RSA.generate(2048)
encrypted_key = key.export_key(passphrase=secret_code, pkcs=8,
                              protection="scryptAndAES128-CBC")

#file_out = open("rsa_key.bin", "wb")
#file_out.write(encrypted_key)
#file_out.close()
print(encrypted_key)
print((key.publickey().export_key()))