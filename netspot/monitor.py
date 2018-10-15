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
import pandas as pd
from scapy.all import sniff
import netifaces
try:  # package mode
    import netspot.stats as stats
    import netspot.utils as utils
    from netspot.recorder import Recorder
    from netspot.sniffer import Sniffer
except BaseException:  # local mode
    import stats
    import utils
    from recorder import Recorder
    from sniffer import Sniffer


class Monitor:
    """
    Main class of the framework. It manages the sniffing, the stats and their computations.
    Moreover, it embeds a logger which records program events (at INFO level) and network
    anomalies (at WARNING level).
    """
    # # Threads
    __monitor_thread = None
    # # Thread events to stop sniffing and monitoring threads
    __stop_monitor_thread_event = threading.Event()
    # config parameters
    config_keys = ["interval",
                   "record_file",
                   "source_type",
                   "source",
                   "sniffing_filter"]

    # the logger
    __logger = logging.getLogger('netspot')

    def __init__(self, interval):
        self.interval = interval
        self.__sniffer = Sniffer()
        self.__recorder = Recorder()
        self.__statistics = []
        self.__values = []
        self.__begin = 0

    def __del__(self):
        self.__stop_monitor_thread_event.set()
        self.__statistics.clear()
        self.__values.clear()

    def _init_recorder(self):
        header = [stat.name for stat in self.__statistics]
        formatter = [stat.fmt for stat in self.__statistics]
        self.__recorder.init_formatters(header, formatter)

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
        """
        Return a Monitor object from a config file
        """
        if os.path.isfile(file_path):
            # create the config parser
            config = configparser.ConfigParser()
            # open the file
            config.read(file_path)
            # basic configuration
            basic_config = config['config']
            # we get the interval
            interval = basic_config.getfloat('interval', 2.0)
            obj = Monitor(interval=interval)
            # we get the output records file
            record_file = basic_config.get('record_file', '/tmp/netspot.csv')
            obj.set_record_file(record_file)
            # type of the source
            source_type = basic_config.get('source_type', 'iface')
            # real source
            source = basic_config.get('source', 'all')
            obj.set_source(source_type, source)
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
                    # we retrive the spot configuration from the .ini file
                    # section
                    spot_config = Monitor._stat_section_to_spot_config(
                        config[s])
                    # we check if this class requires parameters or not
                    if stats.is_requiring_parameters(stat_class):
                        # if it needs parameters, they are in the 'param' key
                        param = [
                            m for m in map(
                                lambda p: p.strip(),
                                config[s].get('param').split(','))]
                        stat = stat_class(*param, **spot_config)
                    else:
                        # otherwise we can directly create the new instance
                        stat = stat_class(**spot_config)
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
                    # we get and set the [default] spot parameters
                    spot_config = Monitor._stat_section_to_spot_config(
                        config['DEFAULT'])
                    # we create the stat (it does not require parameter)
                    stat = stats.AVAILABLE_STATS[stat_name](**spot_config)
                    # and we load the stat
                    obj.load(stat)
            return obj
        else:
            raise OSError(
                "The config file {} does not exists".format(file_path))

    ###########################################################################
    ###########################################################################
    ###########################################################################
    # STATS STUFF
    ###########################################################################
    ###########################################################################
    ###########################################################################

    def stat_from_name(self, stat_name):
        """
        Get the stat from its name
        """
        for s in self.__statistics:
            if s.name == stat_name:
                return s
        raise ValueError("Unknown statistic")

    def get_loaded_stats(self):
        """
        Return real instanciations of the loaded statistics
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
        return dict(stat.spot.status())

    def full_inspection(self):
        """
        Return a dataframe with all the status of all the monitored stats
        """
        status = []
        for stat in self.__statistics:
            stat_status = stat.spot_status()
            stat_status['statistics'] = stat.name
            status.append(stat_status)
        return pd.DataFrame(status)

    ###########################################################################
    ###########################################################################
    ###########################################################################
    # CONFIG - GETTER - SETTER
    ###########################################################################
    ###########################################################################
    ###########################################################################

    def info(self):
        """
        Return configuration information
        """
        return {"interval": self.interval,
                "record_file": self.get_record_file(),
                "source_type": self.get_source_type(),
                "source": self.get_source(),
                "sniffing_filter": self.get_sniffing_filter()}

    def set_record_file(self, file):
        """
        Change the file where the records are saved
        """
        self.__recorder.set_record_file(file)

    def get_record_file(self):
        """
        Get the file where the records are saved
        """
        return self.__recorder.get_record_file()

    def get_source_type(self):
        """
        Return the type of the source of the incoming packets (iface or file)
        """
        return self.__sniffer.get_source_type()

    def get_source(self):
        """
        Return the name of the source (path of iface name)
        """
        return self.__sniffer.get_source()

    def set_source(self, source_type, source):
        self.__sniffer.set_source(source_type, source)

    def get_sniffing_filter(self):
        return self.__sniffer.get_filter()

    def set_sniffing_filter(self, filter_):
        if not self.is_monitoring():
            self.__sniffer.set_filter(filter_)
        else:
            raise RuntimeError("The monitoring is currently active")

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
                # we append it to our stat list
                self.__statistics.append(stat)
                self.__logger.info('Statistic {} loaded'.format(stat.name))
            else:
                raise ValueError('The statistic {} is already loaded'.format(stat.name))
        else:
            raise ValueError('Unknown statistic')
        # add new counters
        self.__sniffer.load(stat.needs)

    def unload(self, stat):
        """
        Remove a previously loaded statistic
        """
        try:
            self.__statistics.remove(stat)
            self.__logger.info('Statistic {} unloaded'.format(stat.name))
        except ValueError:
            raise ValueError('This statistic is not loaded')

    def unload_from_name(self, stat_name):
        """
        Remove a previously loaded statistic from its [unique] name
        """
        index = self.get_loaded_stat_names().index(stat_name)
        if index >= 0:
            self.__statistics.pop(index)
            # self._init_fmt()
            # self._init_header()
            self.__logger.info('Statistic {} unloaded'.format(stat_name))
        else:
            raise ValueError('This statistic is not loaded')

    def clear(self):
        self.__statistics.clear()

    def reset_counters(self):
        """
        Reset all the counters
        """
        self.__sniffer.reset()

    def reset_all_stats(self):
        """
        Reset the spot instances of the loaded statistics
        """
        for stat in self.__statistics:
            stat.reset()
    
    def reset_buffer_and_recoder(self):
        """
        Re open the buffer if the source type is a file
        """
        if self.get_source_type() == 'file':
            pcap_file = self.get_source()
            self.set_source('file', pcap_file)
            self.__logger.info('The capture file {} has be reloaded'.format(pcap_file))
        record_file = self.get_record_file()
        self.set_record_file(record_file)
        self.__logger.info('The record file {} has be reloaded'.format(record_file))



    def is_sniffing(self):
        """
        Return the status of the sniffing
        """
        try:
            return self.__sniffer.is_sniffing()
        except AttributeError:
            return False

    def _compute(self, stat):
        """
        Compute a statistic (and perform the monitoring if it is activated)
        """
        counter_values = self.__sniffer.get_values(
            [counter.name for counter in stat.needs])
        return stat.compute_and_monitor(*counter_values)

    def compute_all_stats(self):
        """
        Compute all the loaded statistics
        """
        return [self._compute(s) for s in self.__statistics]

    def start(self):
        """
        Launch the monitoring (start sniffing if not already started)
        """
        # If scapy is not sniffing, we try to launch it
        if not self.__sniffer.is_sniffing():
            self.__sniffer.start()

        self.__begin = self.__sniffer.time()

        self._init_recorder()
        # we unset the STOP event
        self.__stop_monitor_thread_event.clear()
        # we reset the counters
        self.reset_counters()
        # we create a new thread and start it
        self.__monitor_thread = threading.Thread(daemon=True,
                                                 target=self._monitorloop)
        self.__monitor_thread.start()
        self.__logger.info('Monitoring started')

    def start_if_not(self):
        """
        Launch the monitoring (start sniffing if not already started)
        """
        if self.is_monitoring():
            raise RuntimeError('The monitoring is currently active')
        else:
            self.start()

    def stop(self):
        """
        Stop the monitoring (but not the sniffing)
        """
        # to stop, we have to active the MONITOR STOP event
        self.__stop_monitor_thread_event.set()
        self.__logger.info('Monitoring stopped')
        self.__sniffer.stop()

    def is_monitoring(self):
        """
        Return the status of the monitoring
        """
        try:
            return self.__monitor_thread.is_alive()
        except AttributeError:
            return False

    def live_on(self):
        self.__recorder.live = True
        print('')

    def live_off(self):
        self.__recorder.live = False

    def is_live_mode_on(self):
        if self.__recorder:
            return self.__recorder.live
        else:
            return False

    def _monitorloop(self):
        """
        Main loop: compute statistics and monitor it
        """
        self.reset_counters()
        self.__recorder.reset()
        while (not self.__stop_monitor_thread_event.is_set()) and self.__sniffer.is_sniffing():
            now = self.__sniffer.time()
            if (now - self.__begin) > self.interval:
                try:
                    # get all the new values
                    self.__stat_values = self.compute_all_stats()
                    # store it
                    self.__recorder.save(now, self.__stat_values)
                except Exception as e:
                    self.__logger.warning(e)
                # reset the counters
                self.reset_counters()
                # set the new time basis
                self.__begin = now
        self.__recorder.live = False
        self.__logger.info('Monitoring stopped')
        utils.print_warning('\nThe monitoring has stopped')
