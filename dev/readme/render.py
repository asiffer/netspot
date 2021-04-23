#!/bin/env python3

import jinja2
import os
import argparse

TEMPLATE = "README.template.md"
README = "README.md"


def find_template() -> str:
    path = os.path.abspath(".")
    while not (TEMPLATE in os.listdir(path)):
        # go up
        path = os.path.normpath(os.path.join(path, ".."))
        if path == "/":
            raise RecursionError("Root folder reached "
                                 "without finding template")
    return os.path.join(path, TEMPLATE)


def render(template_file: str, **kwargs):
    template = jinja2.Template(open(template_file, 'r').read())
    return template.render(**kwargs)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        prog="render.py", description="Format README.md")
    parser.add_argument("--version", type=str, required=True,
                        help="the netspot version")

    args = parser.parse_args()

    template_file = find_template()
    root = os.path.dirname(template_file)

    readme_str = render(template_file, **args.__dict__)
    with open(os.path.join(root, README), 'w') as f:
        f.write(readme_str)
