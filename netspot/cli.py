#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Sat Sep 22 14:58:50 2018

@author: asr
"""

import cmd
import pandas as pd
import netspot.stats as stats
import netspot.utils as utils
import netspot.parsers as parsers
from netspot.utils import colorize
from netspot.monitor import Monitor


# to print logs
#pd.set_option('display.max_colwidth', 500)
#pd.set_option('display.max_columns', None)
#pd.set_option('display.expand_frame_repr', False)
#pd.set_option('max_colwidth', -1)

FLAG = r"""
                  _  _ ___ _____ ___ ___  ___ _____
                 | \| | __|_   _/ __| _ \/ _ \_   _|
                 | .` | _|  | | \__ \  _/ (_) || |
                 |_|\_|___| |_| |___/_|  \___/ |_|
"""


class NetSpotCli(cmd.Cmd):
    intro = colorize(FLAG, 'green', bold=True)
    prompt = colorize("(netspot) # ", 'blue', light=True)

    def __init__(self, config_file=None):
        if config_file:
            self.monitor = Monitor.from_config_file(config_file)
        else:
            self.monitor = Monitor(interval=2.0)
        super(NetSpotCli, self).__init__()

    def parseline(self, line):
        command, args, line = super(NetSpotCli, self).parseline(line)
#        print("line:",line, "cmd:",command, "args:",args)
        if args == '':
            args = []
        if args:
            args = args.rstrip().split(' ')
        return command, args, line

    def emptyline(self):
        pass

    def complete_monitor(self, text, line, begidx, endidx):
        choices = ['start', 'stop', 'status']
        if not text:
            return choices
        else:
            return [c for c in choices if c.startswith(text)]

    def complete_config(self, text, line, begidx, endidx):
        choices = Monitor.config_keys
        if not text:
            return choices
        else:
            return [c for c in choices if c.startswith(text)]

    def complete_stat(self, text, line, begidx, endidx):
        choices = ["load", "unload"]
#        print('\nline:', line)
#        print('text:', text)
        split_line = line.split(' ')
        nb_args = len(split_line)
        if nb_args == 1:
            return choices
        elif nb_args == 2:
            return [c for c in choices if c.startswith(split_line[1])]
        if nb_args == 3:
            if split_line[1] == 'load':
                return [
                    s for s in stats.AVAILABLE_STATS if s.startswith(
                        split_line[2])]
            elif split_line[1] == 'unload':
                return [
                    s for s in self.monitor.get_loaded_stat_names() if s.startswith(
                        split_line[2])]

    def complete_inspect(self, text, line, begidx, endidx):
        choices = self.monitor.get_loaded_stat_names()
        if not text:
            return choices
        else:
            return [c for c in choices if c.startswith(text)]

    def print_parameter(self, param=None):
        info = self.monitor.info()
        if not param:
            for key, value in info.items():
                print('{:>20s}\t{}'.format(key, value))
        else:
            print('{:>20s}\t{}'.format(param, info[param]))

    def print_status(self):
        status = ""
        if self.monitor.is_sniffing():
            status += "{:>10s}\t{}\n".format('Sniffing', utils.ShellColor.OK)
        else:
            status += "{:>10s}\t{}\n".format('Sniffing', utils.ShellColor.NO)

        if self.monitor.is_monitoring():
            status += "{:>10s}\t{}\n".format('Monitoring',
                                             utils.ShellColor.OK)
        else:
            status += "{:>10s}\t{}\n".format('Monitoring',
                                             utils.ShellColor.NO)

        status += "\nLoaded statistics\n"
        for obj in self.monitor.get_loaded_stats():
            status += "{:>20s}\t{}\n".format(obj.name, obj.description)

        print(status)

    def set_parameter_value(self, param, value):
        print('value: {}'.format(value))
        if param == 'interval':
            self.monitor.set_interval(value)
        elif param == 'record_file':
            self.monitor.set_record_file(value)
        elif param == 'log_file':
            self.monitor.set_log_file(value)
        elif param == 'log_socket':
            host, port = value.split(':')
            self.monitor.set_log_socket(host, port)
        elif param == 'log_file_level':
            self.monitor.set_log_file_level(value)
        elif param == 'log_socket_level':
            self.monitor.set_log_file_level(value)
        elif param == 'sniffing_iface':
            self.monitor.set_sniffing_interface(value)
        elif param == 'sniffing_filter':
            self.monitor.set_sniffing_filter(value)
        else:
            raise ValueError('Unknown parameter')

    def print_stats(self):
        loaded_stats = self.monitor.get_loaded_stats()
        loaded_stat_class_names = self.monitor.get_loaded_stat_class_names()
        msg = ""
        for stat_class_name, stat_class in stats.AVAILABLE_STATS.items():
            if stat_class_name in loaded_stat_class_names:
                msg += colorize("{:>20s}\t{}\n".format(stat_class_name,
                                                       stat_class.description), 'violet', light=True)
                if stats.is_requiring_parameters(stat_class):
                    for f in filter(
                            lambda stat_obj: stat_obj.__class__.__name__ == stat_class_name,
                            loaded_stats):
                        msg += colorize("{:>30s} {}\n".format('|',
                                                              f.name), 'violet', light=True)
            else:
                msg += "{:>20s}\t{}\n".format(stat_class_name,
                                              stat_class.description)
        print(msg)

    def do_inspect(self, args):
        """
        Inspect the internal state of the spot instances embedded in the statistics
        """
        try:
            parsed_args = parsers.inspect_parser.parse_args(args)
            stat_name = parsed_args.__getattribute__('<monitored-stat>')
        except SystemExit:
            return 0

        if stat_name:
            try:
                print(self.monitor.stat_from_name(stat_name).spot_status())
            except ValueError as e:
                utils.print_error(e)
        else:
            # status is a pandas Dataframe
            status = self.monitor.full_inspection()
            if parsed_args.full:
                columns = ['statistics',
                           'n',
                           'al_up', 'z_up', 't_up', 'Nt_up', 'ex_up',
                           'al_down', 'z_down', 't_down', 'Nt_down', 'ex_down']
            else:
                columns = ['statistics',
                           'n',
                           'al_up', 'z_up',
                           'al_down', 'z_down']
#            status.style.set_properties(**{'text-align': 'right'})
#            with pd.option_context('display.colheader_justify','right'):
            print(status.fillna('-').to_string(columns=columns,
                                               float_format='{:.4f}'.format,
                                               col_space=7))
#            header = "{:>20s} {:>8s} | ".format('statistics', 'n')
#            header += "{:>8s} {:>8s} {:>8s} {:>8s} {:>8s} | ".format(
#                'al_up', 'z_up', 't_up', 'Nt_up', 'ex_up')
#            header += "{:>8s} {:>8s} {:>8s} {:>8s} {:>8s}".format(
#                'al_down', 'z_down', 't_down', 'Nt_down', 'ex_down')
#            print(header)
#            for s in self.monitor.get_loaded_stats():
#                print("{:>20s} ".format(s.name) + s.row_spot_status())

    def do_stat(self, args):
        """
        Display, Load or Unload statistics
        """
        try:
            parsed_args = parsers.stat_parser.parse_args(args).__dict__
        except SystemExit as e:
            return 0

        if len(args) == 0:
            self.print_stats()
            return 0
        else:
            command = args[0]
            stat_name = parsed_args['<stat>']

        try:
            if command == 'load':
                extra = parsed_args['<extra>']
                stat = stats.stat_from_name(stat_name, extra)
                self.monitor.load(stat)
                utils.print_ok(
                    "The statistics {} has been loaded".format(
                        stat.name))
            elif command == 'unload':
                if stat_name == '*':
                    for obj in self.monitor.get_loaded_stats():
                        self.monitor.unload(obj)
                        utils.print_ok(
                            "The statistics {} has been unloaded".format(
                                obj.name))
                else:
                    self.monitor.unload_from_name(stat_name)
                    utils.print_ok(
                        "The statistics {} has been unloaded".format(
                            stat_name))
        except ValueError as e:
            utils.print_warning(e)
        except TypeError as e:
            utils.print_error(e)

    def do_monitor(self, args):
        """
        Monitoring commands
        """
        try:
            parsed_args = parsers.monitor_parser.parse_args(args)
            command = parsed_args.__getattribute__('<command>')
            # command = parsers.monitor_parser.parse_args(
            #     args).__getattribute__('<command>')
        except SystemExit:
            return 0

        try:
            if command == 'start':
                self.monitor.start_monitor_if_not()
                utils.print_warning('The monitoring is started')
                if parsed_args.live:  # if the option 'live' has been added, we trigger the live mode
                    self.monitor.live = True
                if parsed_args.record:  # if the option 'record' has been added, we trigger the record mode
                    self.monitor.record = True
            elif command == 'status':
                self.print_status()
            elif command == 'stop':
                self.monitor.stop_sniffing()
                utils.print_warning('The monitoring is stopped')
        except PermissionError:
            utils.print_error(
                'Scapy seems not to have the rights to listen on interfaces')

    def do_config(self, args):
        """
        Get or set configuration variables
        """
        try:
            parsed_args = parsers.config_parser.parse_args(args).__dict__
            param = parsed_args['<parameter>']
            value = parsed_args['<value>']
        except SystemExit as e:
            return 0

        if not param:
            self.print_parameter()
        elif (param in Monitor.config_keys):
            if not value:
                self.print_parameter(param)
            else:
                try:
                    self.set_parameter_value(param, value)
                    utils.print_ok(
                        'Parameter {} changed to {}'.format(
                            param, value))
                except (ValueError, RuntimeError) as e:
                    utils.print_error(str(e))
        else:
            utils.print_error('Unknown parameter')

    def do_live(self, args):
        """
        Print computed statistics in live
        """
        if self.monitor.is_monitoring():
            self.monitor._init_header()
            self.monitor.live = True
        else:
            utils.print_error("NetSpot is not monitoring")
    
    def do_record(self, args):
        """
        Save computed statistics to csv file
        """
        try:
            action = parsers.record_parser.parse_args(args).command
        except SystemExit:
            return 0
        
        if self.monitor.is_monitoring():
            if self.monitor.record == (action == 'on'):
                utils.print_warning("Recording mode is already turned {}".format(action))
            else:
                self.monitor.record = (action == 'on')
                utils.print_ok("Recording mode is now turned {}".format(action))
        else:
            utils.print_error("NetSpot is not monitoring")

    def do_log(self, args):
        """
        Print logs
        """
        log = pd.read_csv(self.monitor.get_log_file(),
                          header=None,
                          sep='\t',
                          names=['time', 'level', 'message'],
                          dtype={'level': str, 'message': str},
                          date_parser=pd.to_datetime,
                          parse_dates=['time'])
        with pd.option_context('display.max_colwidth', 100, 'display.max_columns', 15):
            print(log.to_string(index=False,
                                header=False,
                                formatters={'message': '{:<60s}'.format}))
#        with open(file, 'r') as log:
#            print(log.read())
#        pass

    def do_EOF(self, args):
        if self.monitor.live:
            self.monitor.live = False
        else:
            answer = input(colorize('\nLeave netspot ? ([y]/n) ', 'yellow'))
            if answer in ['y', 'Y', '']:
                return True
            else:
                return False

    def do_exit(self, args):
        """
        Exit the program
        """
        answer = input(colorize('\nLeave netspot ? ([y]/n) ', 'yellow'))
        if answer in ['y', 'Y', '']:
            return True
        else:
            return False

    def do_help(self, args):
        """
        Print help
        """
        super(NetSpotCli, self).do_help(' '.join(args))


# MyClass.myMethod.__func__.__doc__
NetSpotCli.do_monitor.__doc__ = parsers.monitor_parser.format_help()
NetSpotCli.do_inspect.__doc__ = parsers.inspect_parser.format_help()
NetSpotCli.do_config.__doc__ = parsers.config_parser.format_help()
NetSpotCli.do_stat.__doc__ = parsers.stat_parser.format_help()
