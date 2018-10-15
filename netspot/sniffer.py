
from scapy.all import sniff, rdpcap
from scapy.utils import PcapReader

import os
import time
import threading
import logging
import netifaces
from multiprocessing import Process, Event, Lock

try:  # package mode
    import netspot.counters as counters
except BaseException:  # local mode
    import counters


class Sniffer:
    # counters
    __counters = {}
    # some basic attributes
    __filter = None
    __iface = None
    __source = None
    # thread
    __sniff_thread = None
    # thread event
    __stop_sniff_thread_event = threading.Event()
    # thread lock
    thread_lock = threading.Lock()
    # logger
    __logger = logging.getLogger('netspot')

    def __init__(self, source_type='iface', source='all', filter_=None):
        """
        Parameters
        ----------
        source_type: str
            ('iface' of 'file') The type of the input source
        source: str
            The name of the input source (a path for a file source or an interface
            for the iface source)
        filter_: str
            tcpdump filter (iface mode only)
        """
        self.set_source(source_type, source)
        if self.__source_type == 'iface' and filter_:
            self.set_filter(filter_)
        self.__time = -1

    def __del__(self):
        self.__stop_sniff_thread_event.set()
        self.__counters.clear()

    def get_source_type(self):
        return self.__source_type

    def get_source(self):
        if self.__source_type == 'file':
            return self.__source.filename
        else:
            return self.__source if self.__source else 'all'

    def set_source(self, source_type, source):
        if source_type == 'iface':
            self.__source_type = source_type
            self._set_iface(source)
        elif source_type == 'file':
            self.__source_type = source_type
            self._set_pcap(source)
        else:
            raise ValueError("The source type must be 'file' of 'iface'")

    def _set_iface(self, iface):
        if not self.is_sniffing():
            if (iface == 'all') or (iface is None):
                self.__source = None
            elif iface in netifaces.interfaces():
                self.__source = iface
            else:
                raise ValueError("Unknown interface")
            self.__source_type = 'iface'
            self.__logger.info('Source set to interface {}'.format(iface))
        else:
            raise RuntimeError("The sniffer is currently active")

    def _set_pcap(self, pcap):
        if not self.is_sniffing():
            if os.path.isfile(pcap):
                self.__source = PcapReader(pcap)
                self.__source_type == 'file'
                self.__logger.info('Source set to file {}'.format(pcap))
            else:
                raise ValueError("Unknown interface")
        else:
            raise RuntimeError("The sniffer is currently active")

    def get_filter(self):
        return self.__filter

    def set_filter(self, filter_):
        if not self.is_sniffing():
            self.__filter = filter_
            self.__logger.info('tcpdump filter set to {}'.format(filter_))
        else:
            raise RuntimeError("The sniffer is currently active")

    def _is_valid_counter(self, counter):
        return isinstance(counter, counters.AbstractCounter)

    def _dispatch(self, pkt):
        self.thread_lock.acquire()
        self.__time = pkt.time
        for counter in self.__counters.values():
            counter.process(pkt)
        self.thread_lock.release()


    def _sniff(self):
        sniff(prn=self._dispatch,
              filter=self.__filter,
              iface=self.__source,
              store=False,
              stop_filter=lambda pkt: self.__stop_sniff_thread_event.is_set())

    def _read_pcap(self):
        while not self.__stop_sniff_thread_event.is_set():
            pkt = self.__source.read_packet()
            if pkt:
                self._dispatch(pkt)
            else:
                self.__stop_sniff_thread_event.set()

    def start(self):
        self.__stop_sniff_thread_event.clear()
        if self.__source_type == 'file':
            target = self._read_pcap
            msg = "Start reading {}".format(self.__source.filename)
        else:
            target = self._sniff
            msg = "Start sniffing on {} interface".format(
                self.__source if self.__source else 'all')

        self.__sniff_thread = threading.Thread(daemon=True,
                                               target=target)
        self.__sniff_thread.start()
        self.__logger.info(msg)

    def stop(self):
        self.__stop_sniff_thread_event.set()
        self.__logger.info('Stop capturing')

    def time(self):
        return self.__time

    def reset(self):
        for counter in self.__counters.values():
            counter.reset()

    def load(self, counter_list):
        for counter in counter_list:
            if self._is_valid_counter(counter):
                if counter not in self.__counters.values():
                    self.__counters[counter.name] = counter
                    self.__logger.info(
                        'Counter {} loaded'.format(
                            counter.name))
            else:
                raise ValueError(
                    "The input {} is not a valid counter".format(counter))

    def unload(self, counter_list):
        for counter in counter_list:
            self.__counters.popitem(counter.name)
            self.__logger.info('Counter {} unloaded'.format(counter.name))

    def is_loaded(self, counter):
        if isinstance(counter, counters.AbstractCounter):
            return (counter in self.__counters)
        else:
            raise TypeError("The input is not a valid counter")

    def is_sniffing(self):
        try:
            return self.__sniff_thread.is_alive()
        except AttributeError:
            return False

    def get_all_values(self):
        self.thread_lock.acquire()
        values = [counter.get() for counter in self.__counters.values()]
        self.thread_lock.release()
        return values

    def get_values(self, counter_name_list):
        self.thread_lock.acquire()
        values = [self.__counters[counter_name].get()
                  for counter_name in counter_name_list]
        self.thread_lock.release()
        return values
