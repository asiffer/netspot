#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Sat Sep 22 14:58:50 2018

@author: asr
"""

import cmd
import inspect
import logging
import pandas as pd
import os

try:  # package mode
    import netspot.stats as stats
    import netspot.utils as utils
    import netspot.parsers as parsers
    from netspot.utils import colorize
    from netspot.monitor import Monitor
except BaseException:  # local mode
    import stats
    import utils
    import parsers
    from utils import colorize
    from monitor import Monitor


FLAG = r"""
                  _  _ ___ _____ ___ ___  ___ _____
                 | \| | __|_   _/ __| _ \/ _ \_   _|
                 | .` | _|  | | \__ \  _/ (_) || |
                 |_|\_|___| |_| |___/_|  \___/ |_|
"""


class NetSpotCli(cmd.Cmd):
    intro = colorize(FLAG, 'green', bold=True)
    prompt = colorize("(netspot) # ", 'blue', light=True)
    doc_header = colorize("Available commands", 'green')
    ruler = ''
    short_help = ''
    logger = logging.getLogger('netspot')

    def __init__(self, config_file=None, log_file='/tmp/netspot.log'):
        """
        Constructor

        Parameters
        ----------
        config_file: str
            file to read to get the config of the monitoring
        log_file: str
            file where to store the logs ('/tmp/netspot.log' by default)
        """
        self._init_logger(log_file)
        if config_file:
            self.monitor = Monitor.from_config_file(config_file)
        else:
            self.monitor = Monitor(interval=2)
        self._gen_help()
        super(NetSpotCli, self).__init__()

    def _init_logger(self, file):
        """
        Create the logger.

        Parameters
        ----------
        file: str
            the path to the log gile to use
        """
        if utils.is_file_ok(file):
            self.logger.setLevel(logging.DEBUG)
            handler = logging.FileHandler(file, mode="w+")
            handler.setFormatter(logging.Formatter(
                '%(asctime)s.%(msecs)03d %(module)s %(levelname)s %(message)s',
                datefmt='%H:%M:%S'))
            self.logger.addHandler(handler)
            self.logger.debug("Logger initialized")
        else:
            raise ValueError("The log file is not valid")

    def _gen_help(self):
        """
        This function generate the help message. It gathers the docs of all the actions
        (i.e. the do_* methods) and store it. After that, their documentation is replaced
        by the parser documentation which is richer.
        """
        def predicate(f): return inspect.ismethod(
            f) and f.__name__.startswith('do_')
        details = {}
        for name, obj in inspect.getmembers(self, predicate):
            details[name[3:]] = obj.__doc__.replace(
                '\n', '').replace('\t', '').strip()
        size = str(max(map(len, details)))
        for key, value in details.items():
            self.short_help += '  ' + utils.ShellColor.WHT + utils.ShellColor.BOLD
            self.short_help += ('{:>' + size + 's}').format(key)
            self.short_help += utils.ShellColor.END + '  ' + value + '\n'
        # Then we change the doc. We use the doc of the parser instead (which
        # is richer)
        NetSpotCli.do_monitor.__doc__ = parsers.monitor_parser.format_help()
        NetSpotCli.do_inspect.__doc__ = parsers.inspect_parser.format_help()
        NetSpotCli.do_config.__doc__ = parsers.config_parser.format_help()
        NetSpotCli.do_stat.__doc__ = parsers.stat_parser.format_help()

    def parseline(self, line):
        command, args, line = super(NetSpotCli, self).parseline(line)
        if args == '':
            args = []
        if args:
            args = args.rstrip().split(' ')
        return command, args, line

    def emptyline(self):
        pass

    def complete_monitor(self, text, line, begidx, endidx):
        choices = ['start', 'stop', 'status', 'reset']
        options = ['live']
        split_line = line.split(None)
        inter = set(split_line).intersection(set(choices))

        if len(inter) == 1:
            command = inter.pop()
            split_line.remove("monitor")
            split_line.remove(command)
            opt = split_line.pop()
            if (command == 'start'):
                if text:
                    return [c for c in options if c.startswith(text)]
                elif opt == '-':
                    return ['-']
                elif opt == '--':
                    return options
            else:
                return None
        elif not text:
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
        split_line = line.split(None)
        inter = set(split_line).intersection(set(choices))

        if len(inter) == 1:
            command = inter.pop()
            if command == 'load':
                if text:
                    return [
                    s for s in stats.AVAILABLE_STATS if s.startswith(text)]
                else:
                    return list(stats.AVAILABLE_STATS.keys())
            elif command == 'unload':
                if text:
                    return [
                    s for s in self.monitor.get_loaded_stat_names() if s.startswith(text)]
                else:
                    return self.monitor.get_loaded_stat_names()
            else:
                return None
        elif not text:
            return choices
        else:
            return [c for c in choices if c.startswith(text)]


        # if nb_args == 1:
        #     return choices
        # elif nb_args == 2:
        #     if text:
        #         return [c for c in choices if c.startswith(text)]
        #     elif 
        # if nb_args >= 2:
        #     if split_line[1] == 'load':
        #         return [
        #             s for s in stats.AVAILABLE_STATS if s.startswith(
        #                 split_line[-1])]
        #     elif split_line[1] == 'unload':
        #         return [
        #             s for s in self.monitor.get_loaded_stat_names() if s.startswith(
        #                 split_line[-1])]

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

        if self.monitor.is_monitoring():
            status += "{:>10s}\t{}\n".format('Monitoring',
                                             utils.ShellColor.OK)
        else:
            status += "{:>10s}\t{}\n".format('Monitoring',
                                             utils.ShellColor.NO)

        status += "\nLoaded statistics\n"
        if self.monitor.is_empty():
            status += utils.italic('\tNothing')
        else:
            name_size = max(map(len, self.monitor.get_loaded_stat_names()))
            fmt = "\t{:>" + str(name_size) + "s}\t{}\n"
            for obj in self.monitor.get_loaded_stats():
                status += fmt.format(obj.name, obj.description)
        print(status)

    def set_parameter_value(self, param, value):
        """
        Set the a new value for a given parameter
        """
        print('value: {}'.format(value))
        if param == 'interval':
            self.monitor.set_interval(value)
        elif param == 'record_file':
            self.monitor.set_record_file(value)
        elif param == 'sniffing_filter':
            self.monitor.set_sniffing_filter(value)
        elif param == 'source':
            if os.path.exists(value):
                source_type = 'file'
            else:
                source_type = 'iface'
            self.monitor.set_source(source_type, value)
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
        elif not self.monitor.is_empty():
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
            print(status.fillna('-').to_string(columns=columns,
                                               float_format='{:.4f}'.format,
                                               col_space=7))
        else:
            utils.print_error("No statistics loaded")

    def do_stat(self, args):
        """
        Display, Load or Unload statistics
        """
        try:
            parsed_args = parsers.stat_parser.parse_args(args).__dict__

        except SystemExit as e:
            return 0

        # if no command is given (only 'stat')
        if len(args) == 0:
            self.print_stats()
            return 0
        else:
            command = args[0]
            stat_list = parsed_args['<stats>']

        try:
            if command == 'load':
                for s in stat_list:
                    stat = stats.stat_from_name(s, *parsed_args['parameters'])
                    try:
                        self.monitor.load(stat)
                        utils.print_ok(
                            "The statistic {} has been loaded".format(
                                stat.name))
                    except ValueError as e:
                        utils.print_warning(e)
            elif command == 'unload':
                if '*' in stat_list:
                    for obj in self.monitor.get_loaded_stats():
                        try:
                            self.monitor.unload(obj)
                            utils.print_ok(
                                "The statistic {} has been unloaded".format(
                                    obj.name))
                        except ValueError as e:
                            utils.print_warning(e)
                else:
                    for s in stat_list:
                        try:
                            self.monitor.unload_from_name(s)
                            utils.print_ok(
                                "The statistic {} has been unloaded".format(s))
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
        except SystemExit:
            return 0

        try:
            if command == 'start':
                self.monitor.start_if_not()
                if self.monitor.is_monitoring():
                    utils.print_warning('The monitoring has started', flush=True)
                if parsed_args.live:  # if the option 'live' has been added, we trigger the live mode
                    self.monitor.live_on()
            elif command == 'status':
                self.print_status()
            elif command == 'stop':
                self.monitor.stop()
                utils.print_warning('The monitoring has stopped', flush=True)
            elif command == 'reset':
                self.monitor.reset_all_stats()
                utils.print_warning("The statistics have been reset (you can verify it with the 'inspect' command")
                self.monitor.reset_buffer_and_recoder()
                
        except PermissionError:
            utils.print_error(
                'Scapy seems not to have the rights to listen on interfaces')
        except RuntimeError as e:
            utils.print_error(e)

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
            self.monitor.live_on()
        else:
            utils.print_error("NetSpot is not monitoring")
    
    def do_log(self, args):
        """
        Print logs
        """
        file_handlers = list(filter(lambda h: isinstance(h, logging.FileHandler),
                                    self.logger.handlers))

        if len(file_handlers)>0:
            log_file = os.path.abspath(file_handlers[0].baseFilename)
            try:
                stream = open(log_file, 'rt')
            except FileNotFoundError:
                utils.print_error('The log file ({}) does not exist'.format(log_file))
            formatted_output = ''
            for line in stream:
                level = line.split(' ')[2]
                if level == 'WARNING':
                    formatted_output += utils.colorize(line, 'yellow')
                elif level == 'DEBUG':
                    formatted_output += utils.colorize(line, 'violet')
                else:
                    formatted_output += line
            print(formatted_output)

        else:
            utils.print_error('No log file defined')

    def do_EOF(self, args):
        """
        Stop the live status if activated, otherwise exit the program
        """
        if self.monitor.is_live_mode_on():
            self.monitor.live_off()
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
        answer = input(colorize('Leave netspot ? ([y]/n) ', 'yellow'))
        if answer in ['y', 'Y', '']:
            return True
        else:
            return False

    def do_help(self, args):
        """
        Print this help
        """
        if len(args)>0:
            super(NetSpotCli, self).do_help(args[0])
        else:
            print(self.short_help)
