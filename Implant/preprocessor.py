import json
import os
import re
import random
import string
import shutil
import zlib


def generate_random_name(length=8):
    """Generates a random name of the given length using letters only."""
    return ''.join(random.choices(string.ascii_letters, k=length))


def update_import_paths(file_content, old_path, new_path):
    """Updates the import path in the given file content."""
    pattern = re.escape(old_path)
    return re.sub(pattern, new_path, file_content)


def get_function_names(file_content):
    """Finds all function definitions in the file content excluding the 'main' function."""
    function_pattern = re.compile(r'func\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(')
    function_names = set(match.group(1) for match in function_pattern.finditer(file_content))
    return {name for name in function_names if name != "main"}


def replace_function_names(file_content, name_map):
    """Replaces all occurrences of function names in the file content according to the name map."""
    for old_name, new_name in name_map.items():
        file_content = re.sub(rf'\b{re.escape(old_name)}\b', new_name, file_content)
    return file_content


def process_go_file(input_path, output_path, old_import_path, new_import_path, global_name_map):
    """Reads, processes, and writes a Go file to a new directory with randomized function names and updated import paths."""
    with open(input_path, 'r') as file:
        content = file.read()

    # Do not modify the package declaration
    package_pattern = re.compile(r'^package\s+([A-Za-z_][A-Za-z0-9_]*)', re.MULTILINE)
    match = package_pattern.search(content)
    if match:
        package_name = match.group(1)

    # Update import paths
    content = update_import_paths(content, old_import_path, new_import_path)

    # Extract and randomize function names
    function_names = get_function_names(content)
    for name in function_names:
        if name not in global_name_map:
            if name[0].isupper():  # Preserve exported status
                global_name_map[name] = generate_random_name().capitalize()
            else:
                global_name_map[name] = generate_random_name().lower()

    # Replace all occurrences of function names using the global name map
    updated_content = replace_function_names(content, global_name_map)

    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    with open(output_path, 'w') as file:
        file.write(updated_content)

    #print(f"Processed {input_path} -> {output_path}: {len(function_names)} functions renamed.")


def process_directory(input_directory, output_directory, old_import_path, new_import_path):
    """Processes all Go files in the specified input directory and outputs them to a new directory."""
    global_name_map = {}  # Store function names across all files

    # First pass: Collect all function names across all files
    for root, _, files in os.walk(input_directory):
        for file_name in files:
            if file_name.endswith('.go'):
                input_path = os.path.join(root, file_name)
                with open(input_path, 'r') as file:
                    content = file.read()
                function_names = get_function_names(content)
                for name in function_names:
                    if name not in global_name_map:
                        if name[0].isupper():  # Preserve exported status
                            global_name_map[name] = generate_random_name().capitalize()
                        else:
                            global_name_map[name] = generate_random_name().lower()

    # Second pass: Apply the renaming to all files
    for root, _, files in os.walk(input_directory):
        for file_name in files:
            if file_name.endswith('.go'):
                input_path = os.path.join(root, file_name)
                relative_path = os.path.relpath(input_path, input_directory)
                output_path = os.path.join(output_directory, relative_path)
                process_go_file(input_path, output_path, old_import_path, new_import_path, global_name_map)


def main():
    """Main entry point of the script."""
    import argparse

    parser = argparse.ArgumentParser(description="Randomize Go function names and update import paths.")
    parser.add_argument('input_directory', nargs='?', default='./base_config', help="Path to the directory containing Go files.")
    parser.add_argument('output_directory', nargs='?', default='./preprocessor', help="Path to the directory where processed files will be saved.")
    parser.add_argument('old_import_path', nargs='?', default='NULL/0x27894365/base_config/', help="The old import path to be replaced.")
    parser.add_argument('new_import_path', nargs='?', default='NULL/0x27894365/preprocessor/', help="The new import path to use.")

    args = parser.parse_args()

    if os.path.exists(args.output_directory):
        shutil.rmtree(args.output_directory)  # Clear the output directory if it exists

    process_directory(args.input_directory, args.output_directory, args.old_import_path, args.new_import_path)

    """
    The makefile uses this to determine if listeners are expecting compression
    If the listener wants compression, then the config is compressed and written as a bin file
    This is then embedded at compile time
    The script outputs 'withComp' to the terminal and this tells the makefile to use the 'withComp' tag
    Otherwise, no compression is used and the config is written as a json file and embedded raw
    If you don't use compression, this config is very simple to extract from the exe
    """

    with open('../Server/s_conf.json', 'r') as file:
        data = json.load(file)

    psk1 = data["keys"][0]["psk1"]
    psk2 = data["keys"][0]["psk2"]
    host = data["host_info"][0]["host"]
    method = data["host_info"][0]["method"]
    port = data["host_info"][0]["port"]
    get_path = data["default-GET"][0]["path"]
    post_path = data["default-POST"][0]["path"]
    is_compression = data["default-GET"][0]["comp"].lower() == "true"

    config = {
        "psk1": psk1,
        "psk2": psk2,
        "host": host,
        "method": method,
        "port": port,
        "get_path": get_path,
        "post_path": post_path
    }

    if is_compression:
        # Convert JSON to string and encode to bytes
        json_data = json.dumps(config).encode('utf-8')

        # Compress the data using zlib
        compressed_data = zlib.compress(json_data)

        # Write the binary compressed data to a file
        with open('preprocessor/shared/config.bin', 'wb') as f:
            f.write(compressed_data)

        print('y')

    else:
        json_data = json.dumps(config)

        # Write the JSON data to file
        with open('preprocessor/shared/config.json', 'w') as f:
            f.write(json_data)

        print("n")



if __name__ == "__main__":
    main()
