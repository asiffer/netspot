#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Wed Sep 26 15:42:33 2018

@author: asr
"""
import argparse
import netspot.utils as utils

MONITOR_PARSER_HELP = """\n\
{:>20}\t start the monitoring
{:>20}\t stop the monitoring
{:>20}\t get the current status
""".format(*map(lambda x: utils.colorize(x, 'white', bold=True),
                ['start', 'stop', 'status']))

PARAMETERS = ['interval',
              'record_file',
              'log_file',
              'log_socket',
              'log_file_level',
              'log_socket_level',
              'sniffing_iface',
              'sniffing_filter']

CONFIG_PARSER_HELP = """\
{:>30}\t time interval (in seconds) for counter aggregation (default: 5)
{:>30}\t csv file to store the live statistics (default: None)
{:>30}\t file to store the logs (default: None)
{:>30}\t destination (IP:PORT) to send the logs (default: None)
{:>30}\t logging level for the file (default: INFO)
{:>30}\t logging level for the socket (default: INFO)
{:>30}\t interface to sniff (default: 'all')
{:>30}\t tcpdump filter to apply (default: None)
""".format(*map(lambda x: utils.colorize(x, 'white', bold=True), PARAMETERS))


monitor_parser = argparse.ArgumentParser(
    prog='monitor',
    formatter_class=argparse.RawTextHelpFormatter,
    description=utils.italic("Monitoring commands"))
monitor_parser.add_argument('<command>',
                            choices=['start', 'stop', 'status'],
                            help=MONITOR_PARSER_HELP)
monitor_parser.add_argument(
    '-l',
    '--live',
    action='store_true',
    dest='live',
    help='activate the live mode (print the live statistics')
monitor_parser.add_argument(
    '-r',
    '--record',
    action='store_true',
    dest='record',
    help='activate the record mode (save the computed statistics')

config_parser = argparse.ArgumentParser(
    prog='config',
    formatter_class=argparse.RawTextHelpFormatter,
    description=utils.italic("Set the value of program parameters"))

config_parser.add_argument('<parameter>', nargs='?', help='see above')
config_parser.add_argument(
    '<value>',
    nargs='?',
    help=CONFIG_PARSER_HELP)


stat_parser = argparse.ArgumentParser(
    prog='stat',
    formatter_class=argparse.RawTextHelpFormatter,
    description=utils.italic("Commands around available statistics"))

stat_subparser = stat_parser.add_subparsers()
load_stat_subparser = stat_subparser.add_parser(
    'load',
    formatter_class=argparse.RawTextHelpFormatter,
    epilog="example:\n  (netspot) # load R_SYN\nThe statistics R_SYN has been loaded\n ")
load_stat_subparser.add_argument(
    '<stat>', help="The name of statistics to load")
load_stat_subparser.add_argument(
    '<extra>',
    nargs='*',
    help="Extra parameters (some statistics may require it)")
unload_stat_subparser = stat_subparser.add_parser(
    'unload',
    formatter_class=argparse.RawTextHelpFormatter,
    epilog="example:\n  (netspot) # unload R_SYN\nThe statistics R_SYN has been unloaded\n ")
unload_stat_subparser.add_argument(
    '<stat>',
    help="The name of statistics to unload (you can use * to unload all statistics")


inspect_parser = argparse.ArgumentParser(
    prog='inspect',
    formatter_class=argparse.RawTextHelpFormatter,
    description="Spot inspection command. If the monitored stat is not given, the command returns the status of all the loaded statistics",
    epilog="example:\n  (netspot) # inspect R_SYN\n ")
inspect_parser.add_argument('<monitored-stat>', nargs='?',
                            help='The name of the monitored stat')
inspect_parser.add_argument('--full',
                            dest='full',
                            action="store_true",
                            help='Print the full status of the spot instances')
record_parser = argparse.ArgumentParser(
    prog='record',
    formatter_class=argparse.RawTextHelpFormatter,
    description="Save the records to a csv file",
    epilog="example:\n  (netspot) # record on\n ")
record_parser.add_argument('command', metavar="<command>", choices=['on', 'off'],
                            help='The mode (activated or not)')
