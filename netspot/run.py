#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Thu Sep 27 14:07:24 2018

@author: asr
"""

import argparse
import logging
try:
    from netspot.cli import NetSpotCli
except BaseException:
    from cli import NetSpotCli


def main():
    cli_parser = argparse.ArgumentParser(prog="netspot")
    cli_parser.add_argument(
        "-c",
        "--config",
        dest='config',
        help="config file",
        type=str)
    cli_parser.add_argument(
        "-l",
        "--log",
        dest='log',
        help="log file",
        default="/tmp/netspot.log",
        type=str)
    args = cli_parser.parse_args()

    cli = NetSpotCli(args.config, args.log)
    cli.cmdloop()
