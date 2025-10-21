from flask import Blueprint, jsonify

example_blueprint = Blueprint('example', __name__)

@example_blueprint.route('/example', methods=["GET"])
def main_route():
    return "This is a main route"

@example_blueprint.route('/example/<username>')
def return_data(username):
    return jsonify({"username": username, "message": f"Welcome, {username}!"})
