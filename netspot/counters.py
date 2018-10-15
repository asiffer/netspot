#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Wed Sep 19 16:00:23 2018

@author: asr
"""

try:
    from netspot.utils import is_valid_address
except BaseException:
    from utils import is_valid_address

from abc import abstractmethod, ABCMeta
from scapy.layers.inet import IP, TCP, ICMP


class AbstractCounter(metaclass=ABCMeta):
    """
    Abstract class defining a basic counter
    """

    def __init__(self):
        self.counter = 0
        self.name = self.__class__.__name__

    def __eq__(self, other):
        return isinstance(other, type(self))

    def __hash__(self):
        return hash(self.name)

    @property
    @abstractmethod
    def layer(self):
        """
        Integer which represents the requested layer
        """
        pass

    @abstractmethod
    def process(self, pkt):
        """
        Main method which process a pkt
        """

    @abstractmethod
    def get(self):
        """
        Method which returns the counter
        """
        return self.counter

    @abstractmethod
    def reset(self):
        """
        Main method which reset the counter
        """
        pass


class _IP(AbstractCounter):
    layer = 1

    def get(self):
        return self.counter

    def process(self, pkt):
        if pkt.haslayer(IP):
            self.counter += 1

    def reset(self):
        self.counter = 0


class _ICMP(AbstractCounter):
    layer = 2

    def get(self):
        return self.counter

    def process(self, pkt):
        if pkt.haslayer(ICMP):
            self.counter += 1

    def reset(self):
        self.counter = 0


class _SYN(AbstractCounter):
    layer = 3

    def get(self):
        return self.counter

    def process(self, pkt):
        if pkt.haslayer(TCP):
            flags = str(pkt[TCP].flags)
            if 'S' in flags:
                self.counter += 1

    def reset(self):
        self.counter = 0


class _ACK(AbstractCounter):
    layer = 3

    def get(self):
        return self.counter

    def process(self, pkt):
        if pkt.haslayer(TCP):
            flags = str(pkt[TCP].flags)
            if 'A' in flags:
                self.counter += 1

    def reset(self):
        self.counter = 0


class _IP_BYTES(AbstractCounter):
    layer = 2

    def get(self):
        return self.counter

    def process(self, pkt):
        if pkt.haslayer(IP):
            self.counter += len(pkt[IP])

    def reset(self):
        self.counter = 0


class _UNIQUE_SRC_ADDR(AbstractCounter):
    layer = 2

    def __init__(self):
        self.counter = set()
        self.name = self.__class__.__name__

    def get(self):
        return len(self.counter)

    def process(self, pkt):
        if pkt.haslayer(IP):
            self.counter.add(pkt[IP].src)

    def reset(self):
        self.counter.clear()


class _UNIQUE_DST_ADDR(AbstractCounter):
    layer = 2

    def __init__(self):
        self.counter = set()
        self.name = self.__class__.__name__

    def get(self):
        return len(self.counter)

    def process(self, pkt):
        if pkt.haslayer(IP):
            self.counter.add(pkt[IP].dst)

    def reset(self):
        self.counter.clear()


class _IP_TO_IP(AbstractCounter):
    layer = 2

    def __init__(self, ip_a, ip_b):
        if is_valid_address(ip_a) and is_valid_address(
                ip_b) and (ip_a != ip_b):
            self.pair = set([ip_a, ip_b])
        else:
            raise ValueError("IP addresses are not all valid")
        self.counter = 0
        self.name = self.__class__.__name__ + '_' + '_'.join(self.pair)

    def __eq__(self, other):
        return isinstance(other, type(self)) and (self.pair == other.pair)

    def get(self):
        return self.counter

    def process(self, pkt):
        if pkt.haslayer(IP):
            if self.pair == set([pkt[IP].src, pkt[IP].dst]):
                self.counter += 1

    def reset(self):
        self.counter = 0
