import os
import re
import random

"""
Simple script that's a WIP for inserting opaque predicates into code before compile time
"""

# List of simple opaque predicates to insert
OPAQUE_PREDICATES = [
    "(x * x >= 0)",  # Always true
    "(x ^ x == 0)",  # Always true
    "(x & x == x)",  # Always true
    "((x * x + 1) % 2 == 1)",  # Always true
    "(sha256.Sum256([]byte('constant'))[0] & 1 == 1)",  # Always false
]

def random_predicate():
    """Randomly choose an opaque predicate."""
    return random.choice(OPAQUE_PREDICATES)

def insert_opaque_predicates(file_content):
    """Insert opaque predicates randomly throughout the file."""
    lines = file_content.split("\n")
    modified_lines = []
    for i, line in enumerate(lines):
        # Randomly decide whether to inject an opaque predicate
        if random.random() < 0.2:  # 20% chance to insert opaque predicate
            opaque = random_predicate()
            modified_lines.append(f"if {opaque} {{ }}")

        modified_lines.append(line)

    return "\n".join(modified_lines)

def process_directory(input_dir, output_dir):
    """Copy all files and inject opaque predicates into .go files."""
    for root, _, files in os.walk(input_dir):
        for file_name in files:
            if file_name.endswith('.go'):
                input_path = os.path.join(root, file_name)
                output_path = os.path.join(output_dir, file_name)

                with open(input_path, 'r', encoding='utf-8') as file:
                    content = file.read()

                modified_content = insert_opaque_predicates(content)

                with open(output_path, 'w', encoding='utf-8') as file:
                    file.write(modified_content)

def main():
    input_dir = "./src"
    output_dir = "./preprocessed"
    os.makedirs(output_dir, exist_ok=True)
    process_directory(input_dir, output_dir)
    print(f"Inserted opaque predicates into files in {output_dir}")

if __name__ == "__main__":
    main()
