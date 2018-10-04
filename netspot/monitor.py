#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Wed Sep 19 17:05:53 2018

@author: asr
"""

import sys
import os
import time
import threading
from concurrent.futures import ThreadPoolExecutor
import configparser
import logging
from logging.handlers import SocketHandler
import pandas as pd
from scapy.all import sniff
import netifaces

import netspot.stats as stats
import netspot.utils as utils




class Monitor:
    """
    Main class of the framework. It manages the sniffing, the stats and their computations.
    Moreover, it embeds a logger which records program events (at INFO level) and network
    anomalies (at WARNING level).
    """
    # Threads
    __monitor_thread = None
    __sniff_thread = None
    # Thread events to stop sniffing and monitoring threads
    __stop_sniff_thread_event = threading.Event()
    __stop_monitor_thread_event = threading.Event()
    _thread_error = None
    # switch to print stats
    live = False
    # switch to export stats to file (or database)
    record = False

    # config parameters
    config_keys = ["interval",
                   "record_file",
                   "log_file",
                   "log_socket",
                   "log_file_level",
                   "log_socket_level",
                   "sniffing_iface",
                   "sniffing_filter"]

    # logging formatter
    __logging_formater = logging.Formatter(
        '%(asctime)s\t%(levelname)s\t%(message)s',
        datefmt='%H:%M:%S')

    def __init__(self, interval, log_file=None, record_file=None):
        # start time (initialization is useless)
        self.begin = time.time()
        # time interval to compute the statistics
        self.interval = interval
        # embedded counters
        self.__counters = {}
        # list of the desired statistics
        self.__statistics = []
        # print parameters
        self.__format_options = {'refresh_header_period': 15,
                                 'refresh_header_counter': 15,
                                 'header': '',
                                 'fmt': ''}
        # for scapy sniff
        self.__sniff_options = {'iface': None,
                                'filter': None}
        # the last computed statistics
        self.__stat_values = []
        # data recorder
        self.__recorder = {'buffer': None, 'data': None, 'chunk_size': 10}
        # initilization of the logger
        self.__logger = logging.getLogger('netspot')
        self.__logger.setLevel(logging.INFO)
        if log_file:
            self.set_log_file(log_file, level=logging.INFO)
        if record_file:
            self.set_record_file(record_file)


    @staticmethod
    def _stat_section_to_spot_config(section):
        if isinstance(section, configparser.SectionProxy):
            return {'q': section.getfloat('q'),
                    'n_init': section.getint('n_init'),
                    'level': section.getfloat('level'),
                    'up': section.getboolean('up'),
                    'down': section.getboolean('down'),
                    'bounded': section.getboolean('bounded'),
                    'max_excess': section.getint('max_excess')}
        else:
            raise TypeError(
                "The input variable must a configparser.SectionProxy")

    @staticmethod
    def from_config_file(file_path):
        if os.path.exists(file_path):
            # create the config parser
            config = configparser.ConfigParser()
            # open the file
            config.read(file_path)
            # basic configuration
            basic_config = config['config']
            # we get the interval
            interval = basic_config.getfloat('interval', 5.0)
            # we get the log file
            log_file = basic_config.get('log_file', '/tmp/netspot.log')
            # we get the output records file
            record_file = basic_config.get('record_file', '/tmp/netspot_records.csv')
            # we create the object
            obj = Monitor(interval=interval, log_file=log_file, record_file=record_file)
            # then we can set the level of the log file
            log_file_level = basic_config.get('log_file_level')
            if log_file_level:
                obj.set_log_file_level(log_file_level)
            # we can add a socket log
            log_socket = basic_config.get('log_socket')
            if log_socket:
                obj.set_log_socket(*log_socket.split(':'))
#                obj.set_log_socket(log_socket)
                # and define its level
                log_socket_level = basic_config.get('log_socket_level')
                if log_socket_level:
                    obj.set_log_socket_level(log_socket_level)
            # scapy sniffing interface
            sniffing_iface = basic_config.get('sniffing_iface', 'all')
            obj.set_sniffing_interface(sniffing_iface)
            # scapy sniffing filter (tcpdump)
            sniffing_filter = basic_config.get('sniffing_filter')
            obj.set_sniffing_filter(sniffing_filter)

            # statistics
            # first we get statistics which have a section (the statistics which
            # require additional parameter can only be defined in a section)
            standalone_stats = filter(
                lambda s: s not in [
                    'config',
                    'default',
                    'statistics'],
                config.sections())
            for s in standalone_stats:
                # we get only the available statistics
                if s in stats.AVAILABLE_STATS:
                    # we get the class
                    stat_class = stats.AVAILABLE_STATS[s]
                    # we check if this class requires parameters or not
                    if stats.is_requiring_parameters(stat_class):
                        # if it needs parameters, they are in the 'param' key
                        param = [
                            m for m in map(
                                lambda p: p.strip(),
                                config[s].get('param').split(','))]
                        stat = stat_class(*param)
                    else:
                        # otherwise we can directly create the new instance
                        stat = stat_class()
                    # we retrive the spot configuration from the .ini file
                    # section
                    spot_config = Monitor._stat_section_to_spot_config(
                        config[s])
                    # we set the spot configuration (take the default profile
                    # if parameter is not set)
                    stat.monitor(restart=True, **spot_config)
                    # then we load the stat
                    obj.load(stat)
            # finally we get all the statistics (which do not require any parameters)
            # not loaded loaded yet. They are in the section 'statistics'
            basic_stats = config['statistics']
            for s in basic_stats:
                # config_parser turns the names of the key into lowercase
                stat_name = s.upper()
                # we check that the stat is loadable
                if (stat_name in stats.AVAILABLE_STATS) and \
                        config['statistics'].getboolean(s) and \
                        (not obj.is_loaded(stat_name)):
                    # we create the stat (it does not require parameter)
                    stat = stats.AVAILABLE_STATS[stat_name]()
                    # we get and set the [default] spot parameters
                    spot_config = Monitor._stat_section_to_spot_config(
                        config['DEFAULT'])
                    stat.monitor(restart=True, **spot_config)
                    # and we load the stat
                    obj.load(stat)
            return obj
        else:
            raise OSError(
                "The config file {} does not exists".format(file_path))

    ###########################################################################
    ###########################################################################
    ###########################################################################
    # RECORDING
    ###########################################################################
    ###########################################################################
    ###########################################################################
    def set_record_file(self, file):
        if utils.is_file_ok(file):
            self.__recorder['buffer'] = open(file, 'w')
        else:
            raise ValueError("The file is not valid")

    def get_record_file(self):
        if self.__recorder['buffer']:
            return self.__recorder['buffer'].name
        else:
            return None

    def _is_record_buffer_new(self):
        return self.__recorder['buffer'].tell() == 0

    def export_records(self):
        # print("Data sent to {}".format(self.get_record_file()))
        self.__recorder['data'].to_csv(self.__recorder['buffer'],
                                       header=self._is_record_buffer_new(),
                                       index=None)
        self._init_dataframe()
    
    def _init_dataframe(self):
        columns = ['time'] + self.get_loaded_stat_names()
        self.__recorder['data'] = pd.DataFrame(columns=columns)

    def save_record(self):
        if self.__recorder['data'] is None:
            self._init_dataframe()
        # we append new data
        self.__recorder['data'].loc[self.__recorder['data'].index.size] = [time.time()] + self.__stat_values
        # print(self.__recorder['data'])
        if self.__recorder['data'].index.size == self.__recorder['chunk_size']:
            self.export_records()
            



        

    ###########################################################################
    ###########################################################################
    ###########################################################################
    # LOGGING
    ###########################################################################
    ###########################################################################
    ###########################################################################
#    def debug(self):
#        ch = logging.StreamHandler(sys.stdout)
#        ch.setLevel(logging.DEBUG)
#        ch.setFormatter(self.__logging_formater)
#        self.__logger.addHandler(ch)

    def _get_file_handler(self):
        """
        Return the file handler
        """
        for h in self.__logger.handlers:
            if isinstance(h, logging.FileHandler):
                return h
        return None

    def _get_socket_handler(self):
        """
        Return the socket handler
        """
        for h in self.__logger.handlers:
            if isinstance(h, SocketHandler):
                return h
        return None

    def _add_file_logger(self, file, level=logging.INFO):
        """
        Add new file handler (log will be written to this file)
        """
        handler = logging.FileHandler(file)
        handler.setLevel(level)
        # Add the Formatter to the Handleris
        handler.setFormatter(self.__logging_formater)
        # Add the Handler to the Logger
        self.__logger.addHandler(handler)
        self.__logger.info('File handler {} added'.format(file))

    def _add_socket_logger(self, host, port, level):
        """
        Add new socket handler (log will be sent to this socket)
        """
        handler = SocketHandler(host, port)
        handler.setLevel(level)
        # Add the Formatter to the Handleris
        handler.setFormatter(self.__logging_formater)
        # Add the Handler to the Logger
        self.__logger.addHandler(handler)
        self.__logger.info('Socket handler {}:{} added'.format(host, port))

    def _remove_file_logger(self):
        """
        Remove file logger (if exists)
        """
        self.__logger.removeHandler(self._get_file_handler())

    def _remove_socket_logger(self):
        """
        Remove socket handler (if exists)
        """
        self.__logger.removeHandler(self._get_socket_handler())

    def set_log_file(self, file, level=logging.INFO):
        """
        Define the log file
        """
        if utils.is_file_ok(file):
            self._remove_file_logger()
            self._add_file_logger(file)
            self.__logger.info('Logging file set to {}'.format(file))
        else:
            raise ValueError("The file is not valid")

    def set_log_socket(self, host, port, level=logging.INFO):
        """
        Define the log socket
        """
        if host == 'localhost':
            host = '127.0.0.1'
        if not isinstance(port, int):
            port = utils.to_int(port, "The given port is not valid")
        if utils.is_valid_address(host) and isinstance(
                port, int) and (
                port > 0):
            self._remove_socket_logger()
            self._add_socket_logger(host, port, level)
            self.__logger.info(
                'Logging socket set to {}:{}'.format(
                    host, port))
        else:
            raise ValueError("The socket is not valid")

    def get_log_file(self):
        """
        Return the log file (if it exists)
        """
        handler = self._get_file_handler()
        if handler:
            return handler.baseFilename
        return None

    def get_log_file_level(self):
        """
        Return the log file level (if it exists)
        """
        handler = self._get_file_handler()
        if handler:
            return handler.level
        return None

    def set_log_file_level(self, level):
        """
        Set the log file level

        Parameters
        ----------
        level: int, str
            The log level, it could be either a positive integer or a logging level like 'INFO', 'WARNING' etc.
            For example 'INFO' corresponds to 20 and 'WARNING' to 30.
        """
        if not (isinstance(level, int) or isinstance(level, str)):
            level = utils.to_int(level, "The level is not valid")
        handler = self._get_file_handler()
        if handler:
            handler.setLevel(level)
            self.__logger.info('Logging file level set to {}'.format(level))
        else:
            raise RuntimeError('The log file handler is not defined')

    def get_log_socket(self):
        """
        Return the log socket address:port
        """
        handler = self._get_socket_handler()
        if handler:
            return "{}:{}".format(*handler.address)
        return None

    def get_log_socket_level(self):
        """
        Return the log socket level (if it exists)
        """
        handler = self._get_socket_handler()
        if handler:
            return handler.level
        return None

    def set_log_socket_level(self, level):
        """
        Set the log socket level.

        Parameters
        ----------
        level: int, str
            The log level, it could be either a positive integer or a logging level like 'INFO', 'WARNING' etc.
            For example 'INFO' corresponds to 20 and 'WARNING' to 30.
        """
        if not (isinstance(level, int) or isinstance(level, str)):
            level = utils.to_int(level, "The level is not valid")

        handler = self._get_socket_handler()
        if handler:
            handler.setLevel(level)
            self.__logger.info('Logging socket level set to {}'.format(level))
        else:
            raise RuntimeError('The log file handler is not defined')

    ###########################################################################
    ###########################################################################
    ###########################################################################
    # LIVE FORMAT
    ###########################################################################
    ###########################################################################
    ###########################################################################
    def _init_header(self):
        """
        Initialize the header string (for live print)
        """
        names = [s.name for s in self.__statistics]
        self.__format_options['header'] = '\t'.join(names)
        self.__format_options['refresh_header_counter'] = \
            self.__format_options['refresh_header_period']
        self.__logger.debug('Header initialized')

    def _init_fmt(self):
        """
        Initialize the format string (for live print)
        """
        fmts = [s.fmt for s in self.__statistics]
        self.__format_options['fmt'] = '\t'.join(fmts)
        self.__logger.debug('Format string initialized')

    ###########################################################################
    ###########################################################################
    ###########################################################################
    # SNIFF OPTIONS
    ###########################################################################
    ###########################################################################
    ###########################################################################
    def set_sniffing_interface(self, iface):
        if iface == 'all':
            self.__sniff_options['iface'] = None
        elif iface in netifaces.interfaces():
            self.__sniff_options['iface'] = iface
        else:
            raise ValueError(
                "The network interface {} does not exist".format(iface))

    def set_sniffing_filter(self, tcpdump_filter):
        if tcpdump_filter is None:
            self.__sniff_options['filter'] = None
        elif isinstance(tcpdump_filter, str):
            self.__sniff_options['filter'] = tcpdump_filter
        else:
            raise ValueError("The filter must be a string")

    ###########################################################################
    ###########################################################################
    ###########################################################################
    # MISC
    ###########################################################################
    ###########################################################################
    ###########################################################################
    def stat_from_name(self, stat_name):
        for s in self.__statistics:
            if s.name == stat_name:
                return s
        raise ValueError("Unknown statistics")

    def get_thread_error(self):
        return self._thread_error

    def get_loaded_stats(self):
        """
        Return real instanciations of the loadded statistics
        """
        return list.copy(self.__statistics)

    def get_loaded_stat_names(self):
        """
        Return the corresponding names of the loaded statistics
        """
        return [s.name for s in self.__statistics]

    def get_loaded_stat_classes(self):
        """
        Return the corresponding classes of the loaded statistics
        """
        return [s.__class__ for s in self.__statistics]

    def get_loaded_stat_class_names(self):
        """
        Return the names of the classes of the loaded statistics [lol]
        """
        return [s.__class__.__name__ for s in self.__statistics]

    def get_available_stats(self):
        """
        Return the names of the available statistics
        """
        return list(stats.AVAILABLE_STATS)

    def is_empty(self):
        """
        Return True if there is not any loaded stats
        """
        return len(self.__statistics) == 0

    def inspect(self, stat_name):
        """
        Return the status of the spot instance monitoring the feature
        """
        stat = self.stat_from_name(stat_name)
        if stat.monitored:
            return dict(stat.spot.status())
        else:
            raise ValueError(
                "The statistics {} is not monitored".format(stat_name))

    def full_inspection(self):
        """
        Return a dataframe with all the status of all the monitored stats
        """
        status = []
        for stat in self.__statistics:
            if stat.monitored:
                stat_status = stat.spot_status()
                stat_status['statistics'] = stat.name
                status.append(stat_status)
        return pd.DataFrame(status)

    def info(self):
        """
        Return configuration information
        """
        return {"interval": self.interval,
                "record_file": self.get_record_file(),
                "log_file": self.get_log_file(),
                "log_socket": self.get_log_socket(),
                "log_file_level": self.get_log_file_level(),
                "log_socket_level": self.get_log_socket_level(),
                "sniffing_iface": self.__sniff_options['iface']
                if self.__sniff_options['iface'] else 'all',
                "sniffing_filter": self.__sniff_options['filter']}

    def set_interval(self, value):
        """
        Set the value of the interval aggregation
        """
        if not (isinstance(value, int) or isinstance(value, float)):
            try:
                value = float(value)
            except ValueError:
                raise ValueError("The value must be real and positive")
        if value > 0:
            self.interval = value
            self.__logger.info('Interval set to {}'.format(value))
        else:
            raise ValueError("The value must be real and positive")

    def is_loaded(self, stat_name):
        """
        Check if the stat (through its name) is already loaded
        """
        return (stat_name in self.get_loaded_stat_names())

    def load(self, stat):
        """
        Load a new statistic
        """
        if isinstance(stat, stats.AtomicStat):
            if stat not in self.__statistics:
                # we create a spot instance to monitor it
                if not stat.monitored:
                    stat.monitor()
                # we append it to our stat list
                self.__statistics.append(stat)

                self.__logger.info('Statistics {} loaded'.format(stat.name))
            else:
                raise ValueError('This statistics is already loaded')
        else:
            raise ValueError('Unknown statistics')

        # add new counters
        for _need in stat.needs:
            if _need.name not in self.__counters:
                self.__counters[_need.name] = _need

        # update the header and the format string
        self._init_fmt()
        self._init_header()

    def unload(self, stat):
        """
        Remove a previously loaded statistic
        """
        try:
            self.__statistics.remove(stat)
            self._init_fmt()
            self._init_header()
            self.__logger.info('Statistics {} unloaded'.format(stat.name))
        except ValueError:
            raise ValueError('This statistics is not loaded')

    def unload_from_name(self, stat_name):
        """
        Remove a previously loaded statistic from its [unique] name
        """
        index = self.get_loaded_stat_names().index(stat_name)
        if index >= 0:
            self.__statistics.pop(index)
            self._init_fmt()
            self._init_header()
            self.__logger.info('Statistics {} unloaded'.format(stat_name))
        else:
            raise ValueError('This statistics is not loaded')

    def reset_counters(self):
        """
        Reset all the counters
        """
        for _counter in self.__counters.values():
            _counter.reset()
        self.begin = time.time()
        self.__logger.info('Counters reset')

    def is_sniffing(self):
        """
        Return the status of the sniffing
        """
        try:
            return self.__sniff_thread.is_alive()
        except AttributeError:
            return False

    def _sniff(self):
        """
        Scapy sniff function
        """
        try:
            sniff(
                prn=self._process,
                filter=self.__sniff_options['filter'],
                iface=self.__sniff_options['iface'],
                store=False,
                stop_filter=lambda pkt: self.__stop_sniff_thread_event.is_set())
        except BaseException:
            self.__stop_sniff_thread_event.set()
            _, self._thread_error, _ = sys.exc_info()

    def _process(self, pkt):
        """
        Send the packet to the counters so as to let them process it
        """
        for _counter in self.__counters.values():
            _counter.process(pkt)

    def _compute(self, stat):
        """
        Compute a statistic (and perform the monitoring if it is activated)
        """
        counter_values = [self.__counters[n.name].get() for n in stat.needs]
        return stat.compute_and_monitor(*counter_values)

    def compute_all_stats(self):
        """
        Compute all the loaded statistics
        """
        return [self._compute(s) for s in self.__statistics]
        # with ThreadPoolExecutor() as executor:
        #     futures = [executor.submit(self._compute, s)
        #                for s in self.__statistics]
        # return [f.result() for f in futures]

    def get_counters_for(self, stat):
        """
        Get the values of the counters required to compute the given stat
        """
        return [self.__counters[n.name].get() for n in stat.needs]

    def start_monitor(self):
        """
        Launch the monitoring (start sniffing if not already started)
        """
        # If sscapy is not sniffing, we try to launch it
        if not self.is_sniffing():
            self.start_sniffing()
            # we need a certain amount of time to be sure that nothing happened
            time.sleep(0.1)
        # if scapy fails
        if self._thread_error:
            # we get the exception
            exception = self._thread_error
            # we reset it
            self._thread_error = None
            # and we raise it
            raise exception
        else:
            # we unset the STOP event
            self.__stop_monitor_thread_event.clear()
            # we reset the counters
            self.reset_counters()
            # we create a new thread and start it
            self.__monitor_thread = threading.Thread(daemon=True,
                                                     target=self._monitorloop)
            self.__monitor_thread.start()
            self.__logger.info('Monitoring started')

    def start_monitor_if_not(self):
        """
        Launch the monitoring (start sniffing if not already started)
        """
        if self.is_monitoring():
            raise RuntimeError('The monitoring is currently active')
        else:
            self.start_monitor()

    def stop_monitor(self):
        """
        Stop the monitoring (but not the sniffing)
        """
        # to stop, we have to active the MONITOR STOP event
        self.__stop_monitor_thread_event.set()
        self.__logger.info('Monitoring stopped')

    def start_sniffing(self):
        """
        Start to sniff the network
        """
        self.__stop_sniff_thread_event.clear()
        self.__sniff_thread = threading.Thread(daemon=True,
                                               target=self._sniff)

        self.__sniff_thread.start()

    def stop_sniffing(self):
        """
        Stop to sniff network (and monitoring too if started)
        """
        # to stop, we have to active the MONITOR STOP event
        self.__stop_sniff_thread_event.set()
        self.stop_monitor()
        self.__logger.info('Sniffing stopped')

    def is_monitoring(self):
        """
        Return the status of the monitoring
        """
        try:
            return self.__monitor_thread.is_alive()
        except AttributeError:
            return False

    def _monitorloop(self):
        """
        Main loop: compute statistics and monitor it
        """
        self.reset_counters()
        while not self.__stop_monitor_thread_event.is_set():
            if (time.time() - self.begin) > self.interval:
                # get all the new values
                self.__stat_values = self.compute_all_stats()
                # retrieve the logs
                self._log_spot_output()
                # print values if live mode
                if self.live:
                    print(self.print_values())
                if self.record:
                    self.save_record()
                # reset counters
                self.reset_counters()

    def _log_spot_output(self):
        """
        Log the monitoring result (if an event occurs)
        """
        for stat, stat_val in zip(self.__statistics, self.__stat_values):
            if stat.last_spot_status == 1:  # UP alarm
                self.__logger.warning(
                    '%15s Alarm [value: %.3f, p: %.3e]',
                    '[' + stat.name + ']',
                    stat_val,
                    stat.spot.up_probability(stat_val))
            elif stat.last_spot_status == -1:  # DOWN alarm
                self.__logger.warning(
                    '%15s Alarm [value: %.3f, p: %.3e]',
                    '[' + stat.name + ']',
                    stat_val,
                    stat.spot.up_probability(stat_val))
            elif stat.last_spot_status == 4:  # Calibration
                self.__logger.info(
                    '%15s Calibration',
                    '[' + stat.name + ']')

    # def _log_spot_output(self, index):
    #     """
    #     Log the monitoring result (if an event occurs)
    #     """
    #     stat = self.__statistics[index]
    #     stat_val = self.__stat_values[index]
    #     if stat.last_spot_status == 1:  # UP alarm
    #         self.__logger.warning(
    #             '%15s Alarm [value: %.3f, p: %.3e]',
    #             '[' + stat.name + ']',
    #             stat_val,
    #             stat.spot.up_probability(stat_val))
    #     elif stat.last_spot_status == -1:  # DOWN alarm
    #         self.__logger.warning(
    #             '%15s Alarm [value: %.3f, p: %.3e]',
    #             '[' + stat.name + ']',
    #             stat_val,
    #             stat.spot.up_probability(stat_val))
    #     elif stat.last_spot_status == 4:  # Calibration
    #         self.__logger.info(
    #             '%15s Calibration',
    #             '[' + stat.name + ']')

    def _name_of(self, index):
        """
        Return the name of a statistics given its index
        """
        return self.__statistics[index].name

    def print_values(self):
        """
        Print current statistics values
        """
        # if we have to refresh the header
        if self.__format_options['refresh_header_counter'] == \
                self.__format_options['refresh_header_period']:
            # we skip a line and rewrite the header
            msg = '\n' + self.__format_options['header'] + '\n'
            # reset the counter
            self.__format_options['refresh_header_counter'] = 0
        else:
            msg = ""
            # otherwise, we will write new values so the counter is incremented
            self.__format_options['refresh_header_counter'] += 1
        # we write all the stats
        msg += self.__format_options['fmt']
        return msg.format(*self.__stat_values)
