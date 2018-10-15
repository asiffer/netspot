#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Wed Sep 26 15:42:33 2018

@author: asr
"""
import argparse
import netspot.utils as utils


MONITOR_PARSER_HELP = """\n
{4:}{0:>20}{5:}\t start the monitoring
{4:}{1:>20}{5:}\t stop the monitoring
{4:}{2:>20}{5:}\t get the current status
{4:}{3:>20}{5:}\t reset the spot instances
""".format('start',
           'stop',
           'status',
           'reset',
           utils.ShellColor.BOLD + utils.ShellColor.WHT,
           utils.ShellColor.END)

MONITOR_EPILOG = """
example:
  (netspot) # stat load R_ICMP R_SYN
  The statistics R_ICMP has been loaded
  The statistics R_SYN has been loaded
  (netspot) # monitor start
  The monitoring has started
  (netspot) # monitor status
  Monitoring      [OK]

  Loaded statistics
          R_ICMP  Ratio of ICMP packets
           R_SYN  Ratio of SYN packets"""

CONFIG_PARSER_HELP = """\n
{4:}{0:>20s}{5:}\t time interval (in seconds) for counter aggregation (default: 2)
{4:}{1:>20s}{5:}\t csv file to store the live statistics (default: None)
{4:}{2:>20s}{5:}\t input source (default: 'all')
{4:}{3:>20s}{5:}\t tcpdump filter to apply (default: None)
""".format('interval',
           'record_file',
           'source',
           'sniffing_filter',
           utils.ShellColor.BOLD + utils.ShellColor.WHT,
           utils.ShellColor.END)


monitor_parser = argparse.ArgumentParser(
    prog='monitor',
    formatter_class=argparse.RawTextHelpFormatter,
    description=utils.italic("Monitoring commands") + MONITOR_PARSER_HELP,
    epilog=MONITOR_EPILOG)
monitor_parser.add_argument('<command>',
                            choices=['start', 'stop', 'status', 'reset'],
                            help="See above")
monitor_parser.add_argument(
    '-l',
    '--live',
    action='store_true',
    dest='live',
    help='activate the live mode (print the live statistics)')

config_parser = argparse.ArgumentParser(
    prog='config',
    formatter_class=argparse.RawTextHelpFormatter,
    description=utils.italic("Set the value of program parameters") +
    CONFIG_PARSER_HELP)

config_parser.add_argument('<parameter>', nargs='?', help='see above')
config_parser.add_argument(
    '<value>',
    nargs='?',
    help='The value of the parameter')


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
    '<stats>', nargs='+', help="The name of the statistics to load")
load_stat_subparser.add_argument(
    '-p',
    '--parameters',
    nargs='+',
    default=[],
    help="Extra parameters (some statistics may require it)")
unload_stat_subparser = stat_subparser.add_parser(
    'unload',
    formatter_class=argparse.RawTextHelpFormatter,
    epilog="example:\n  (netspot) # unload R_SYN\nThe statistics R_SYN has been unloaded\n ")
unload_stat_subparser.add_argument(
    '<stats>',
    nargs='+',
    default=[],
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
record_parser.add_argument(
    'command', metavar="<command>", choices=[
        'on', 'off'], help='The mode (activated or not)')
