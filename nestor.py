import argparse
import os
import json
from pathlib import Path
import shutil
import hashlib

HERE = Path.cwd()
NESTOR_PATH = Path(".ne/store")

def init():
    # TODO: probably check that if it exists, it is a directory
    if not Path.exists(NESTOR_PATH):
        # creates ./.ne/store
        Path.mkdir(NESTOR_PATH, parents=True)

def search_ne_store_up(here):
    if Path.exists(here.joinpath(NESTOR_PATH)):
        return here.joinpath(NESTOR_PATH)
    if here == Path("/"):
        return None
    return search_ne_store_up(here.parent)

def update(target_file, inputs_dict):
    target_path = Path(target_file)

    # 1. Resolve the symlink
    result_file = target_path.resolve()

    # 2. Find the previous json and result
    previous_inputs_json = result_file.parent.joinpath(Path("nestor.json"))
    with open(previous_inputs_json, "r") as f:
        previous_inputs = json.loads(f.read())

    # 3. Join the json
    previous_inputs.update(inputs_dict)

    # 4. Add the new path to the store
    add(result_file, previous_inputs, target_file, move=False)


def add(result_file, inputs_dict, symlink_name, move=True):
    nestor_path = search_ne_store_up(Path.cwd())
    if nestor_path is None:
        print("Could not find ne/store")
        return -1
    result_path = Path(result_file)

    # 1. Hash all the inputs
    inputs_hashed = {}
    for (k, v) in inputs_dict.items():
        p = Path(v)
        h = -1
        if Path.exists(p):
            with open(p, "rb") as f:
                h = hashlib.file_digest(f, "sha1").hexdigest()
        else:
            h = hashlib.sha1(v)
        inputs_hashed[k] = h

    # 2. Hash the resulting json
    global_hash = hashlib.sha1(json.dumps(inputs_hashed, sort_keys=True).encode('utf-8')).hexdigest()

    # 3. Create folder
    output_folder_path = nestor_path.joinpath(Path(f"{global_hash}-{result_path.name}"))
    output_path = output_folder_path.joinpath(Path(result_path.name))
    if not Path.exists(output_folder_path):
        Path.mkdir(output_folder_path, parents=False)

        # 4. Store the result and the original json (sorted)
        if move:
            shutil.move(result_path, output_path) 
        else:
            shutil.copy(result_path, output_path)
        with open(output_folder_path.joinpath(Path("nestor.json")), "w") as nestor_json:
            nestor_json.write(json.dumps(inputs_dict, sort_keys=True))
    else:
        print("Already stored!")

    # 5. set up symlink
    symlink_path = Path(symlink_name)
    if Path.exists(symlink_path):
        symlink_path.unlink()
    relative_output_path = os.path.relpath(output_path, symlink_path.parent)
    symlink_path.symlink_to(relative_output_path)
    return 0

def main():
    parser = argparse.ArgumentParser(prog="Nestor")
    subparsers = parser.add_subparsers(dest="subcommand", help="sub-command help")

    # INIT
    parser_init = subparsers.add_parser("init", help="init help")

    # ADD
    parser_add = subparsers.add_parser("add", help="add help")
    parser_add.add_argument("-d", "--deps", nargs="+", help="bar help")
    parser_add.add_argument("result", help="bar help")

    # UPDATE
    parser_update = subparsers.add_parser("update", help="update help")
    parser_update.add_argument("-d", "--deps", nargs="+", help="bar help")
    parser_update.add_argument("target", help="bar help")


    args = parser.parse_args()
    command = args.subcommand
    if command == "init":
        init()
    elif command == "add":
        inputs_dict = dict(map(lambda x: x.split(":"), args.deps))
        add(args.result, inputs_dict, args.result)
    elif command == "update":
        inputs_dict = dict(map(lambda x: x.split(":"), args.deps))
        update(args.target, inputs_dict)
    else:
        print(f"'{command}' not supported yet")
        print(args)
    return 0

main()
