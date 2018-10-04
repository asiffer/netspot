#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Thu Sep 27 14:07:24 2018

@author: asr
"""

import argparse
from netspot.cli import NetSpotCli

# if __name__ == '__main__':


def main():
    cli_parser = argparse.ArgumentParser(prog="netspot")
    cli_parser.add_argument(
        "-c",
        "--config",
        dest='config',
        help="config file",
        type=str)
    args = cli_parser.parse_args()
    cli = NetSpotCli(args.config)
    cli.cmdloop()
