import gc

import bcrypt
import pickle
import getpass


# Playing with type hinting in these functions
def compare_hash(operator: str, pw_to_compare: str) -> bool:
    """
    Compares a password and a hash
    Pulls them from a pkl file stored locally
    :param operator: The username to get the associated hash
    :param pw_to_compare: The password being compared
    :return: Bool indicating pass or fail
    """
    with open('users.pkl', 'rb') as file:
        # Dump the data into a local dict
        data = pickle.load(file)

    # Search the dict with the operator ID as the search key
    if data[operator] and operator in data:
        # Use bcrypt's built in method to check the password
        pw_comp_result = bcrypt.checkpw(bytes(pw_to_compare, 'utf-8'), data[operator])
        if pw_comp_result is True:
            # Delete the sensitive data
            del pw_to_compare, data, operator, pw_comp_result
            # Garbage collect
            gc.collect()
            return True
        else:
            # Delete the sensitive data
            del pw_to_compare, data, operator, pw_comp_result
            # Garbage collect
            gc.collect()
            return False
    else:
        # Delete the sensitive data
        del pw_to_compare, data, operator
        # Garbage collect
        gc.collect()
        return False


def save_password(inpt_uname: str, inpt_pw) -> bool:
    """
    Add a user password to the store
    :param inpt_uname: The username to associate with the password
    :param inpt_pw: The users' password
    :return: Bool indicating pass or fail
    """
    # Empty dict
    data = {}
    # Open the json file for users
    try:
        with open('users.pkl', 'rb') as file:
            # Dump the data into a local dict
            data = pickle.load(file)
    except Exception as e:
        print(f"No users exist yet: {e}")
        # If the dict is empty this will cause issues
        # First user creation needs to be handled specially
        salt = bcrypt.gensalt()
        hashedpw = bcrypt.hashpw(bytes(inpt_pw, 'utf-8'), salt)
        data[inpt_uname] = hashedpw

        # Save the new data
        with open('users.pkl', 'wb') as file:
            # Dump the data into a local dict
            pickle.dump(data, file)
            # Delete the sensitive data
            del inpt_pw, hashedpw, data
            # Garbage collect
            gc.collect()
            return True

    # Check the operator name doesn't exist
    if inpt_uname not in data:
        salt = bcrypt.gensalt()
        hashedpw = bcrypt.hashpw(bytes(inpt_pw, 'utf-8'), salt)
        data[inpt_uname] = hashedpw

        with open('users.pkl', 'wb') as file:
            # Write the data to the file
            pickle.dump(data, file)
            # Delete the sensitive data
            del inpt_pw, hashedpw, data
            # Garbage collect
            gc.collect()
            return True

    else:
        # Delete the sensitive data
        del inpt_pw, data
        # Garbage collect
        gc.collect()
        return False


if __name__ == "__main__":
    uname = input("Enter username to register:" )
    pw = getpass.getpass()
    result = save_password(uname, pw)
    if result is True:
        print("Registered user")
    else:
        print("Name exists, choose a different one")
