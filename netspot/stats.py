#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Wed Sep 19 16:37:58 2018

@author: asr
"""


import pylibspot

from abc import abstractmethod, ABCMeta
import sys
import inspect
import logging

try:
    import netspot.counters as counters
    import netspot.utils as utils
except BaseException:
    import counters
    import utils

###############################################################################
###############################################################################
###############################################################################
# ABSTRACT BASE CLASS
###############################################################################
###############################################################################
###############################################################################


class AtomicStat(metaclass=ABCMeta):
    """
    Abstract class defining a basic statistic which can be computed through
    counters
    """
    __logger = logging.getLogger('netspot')
    __spot_status = None
    __value = None

    __row_spot_status_format = "{:>8d} {:>8.3f} {:>8.3f} {:>8d} {:>8d}"
    __spot_status_format = "\
{0:>8s}  {5:>8d}  Number of *** alerts triggered\n\
{1:>8s}  {6:>8.3f}  Value of the *** decision threshold\n\
{2:>8s}  {7:>8.3f}  Value of the *** transitional threshold\n\
{3:>8s}  {8:>8d}  Total number of observed *** peaks\n\
{4:>8s}  {9:>8d}  Current number of stored *** peaks (for the fit)"
#         n  {:>8d}  Total number of normal observations

    def __init__(self, **kwargs):
        """
        **kwargs
            SPOT algorithm parameters
        """
        self.name = self.__class__.__name__
        self.__spot = pylibspot.Spot(**kwargs)

    def counter_names(self):
        return [n.name for n in self.needs]

    @property
    @abstractmethod
    def description(self):
        """
        Description of the statistic
        """
        pass

    @property
    @abstractmethod
    def needs(self):
        """
        Required counters to compute the statistic
        """
        pass

    @property
    @abstractmethod
    def fmt(self):
        """
        "printf" format to display the value
        ex: {:.2f}
        """
        pass

    @abstractmethod
    def compute(self, *args):
        """
        Main method which computes the statistic from the given counters
        """
        pass

    def compute_and_monitor(self, *args):
        self.__value = self.compute(*args)
        self.__spot_status = self.__spot.step(self.__value)
        self._log()
        return self.__value

    def _log(self):
        if self.__spot_status == 1:  # UP alarm
            self.__logger.warning(
                '%15s Alarm [value: %.3f, p: %.3e]',
                '[' + self.name + ']',
                self.__value,
                self.__spot.up_probability(self.__value))
        elif self.__spot_status == -1:  # DOWN alarm
            self.__logger.warning(
                '%15s Alarm [value: %.3f, p: %.3e]',
                '[' + self.name + ']',
                self.__value,
                self.__spot.down_probability(self.__value))
        elif self.__spot_status == 4:  # Calibration
            self.__logger.info(
                '%15s Calibration',
                '[' + self.name + ']')

    def reset(self, **kwargs):
        """
        Reset the spot instance
        """
        # current config
        config = dict(self.__spot.config())
        # update the config is new parameters are given
        config.update(kwargs)
        # recreate the object
        self.__spot = pylibspot.Spot(**config)
        # log it
        self.__logger.info('Statistics {} reset'.format(self.name))

    def __eq__(self, other):
        if isinstance(other, AtomicStat):
            return self.name == other.name
        else:
            raise TypeError("We can only compare AtomicStats")

    def spot_status(self):
        status = dict(self.__spot.status())
        config = dict(self.__spot.config())
        # if up is not set, we remove the linked values
        if not config['up']:
            for k in filter(lambda k: '_up' in k, status.keys()):
                status[k] = None
        # if down is not set, we remove the linked values
        if not config['down']:
            for k in filter(lambda k: '_down' in k, status.keys()):
                status[k] = None
        if status['n'] < config['n_init']:  # if not initialized
            for k in ['z_up', 'z_down', 't_up', 't_down']:
                status[k] = None
        return status

    def row_spot_status(self):
        """
        Row formatting of the spot status
        """
        conf = dict(self.__spot.config())
        status = dict(self.__spot.status())
        output = "{:>8d} | ".format(status['n'])
        if conf['up']:
            output += self.__row_spot_status_format.format(status['al_up'],
                                                           status['z_up'],
                                                           status['t_up'],
                                                           status['Nt_up'],
                                                           status['ex_up'])
        else:
            output += ' ' * 44
        output += ' | '
        if conf['down']:
            output += self.__row_spot_status_format.format(
                status['al_down'],
                status['z_down'],
                status['t_down'],
                status['Nt_down'],
                status['ex_down'])
        else:
            output += ' ' * 44
        return output

    def str_spot_status(self):
        """
        Fancy formatting of the spot status
        """

        conf = dict(self.__spot.config())
        status = dict(self.__spot.status())
        base_header = ['al_***', 'z_***', 't_***', 'Nt_***', 'ex_***']
        output = "\n{:>8s}  {:>8d}  Total number of normal observations\n".format(
            'n', status['n'])
        if conf['up']:
            up_base = [b.replace('***', 'up') for b in base_header]
            filled_up_base = up_base + [status[b] for b in up_base]
            output += self.__spot_status_format.format(
                *filled_up_base).replace('***', 'up') + '\n'
        if conf['down']:
            down_base = [b.replace('***', 'down') for b in base_header]
            filled_down_base = down_base + [status[b] for b in down_base]
            output += self.__spot_status_format.format(
                *filled_down_base).replace('***', 'down') + '\n'
        return output


###############################################################################
###############################################################################
###############################################################################
# SPECIFIC IMPLEMENTATIONS (STATISTICS)
###############################################################################
###############################################################################
###############################################################################


class R_SYN(AtomicStat):
    """
    Ratio of SYN packets
    """

    description = "Ratio of SYN packets"
    needs = [counters._SYN(), counters._IP()]
    fmt = "{:.3f}"

    def compute(self, syn, ip):
        if syn == 0:
            return 0
        return 100. * syn / ip


class R_ACK(AtomicStat):
    """
    Ratio of ACK packets
    """

    description = "Ratio of ACK packets"
    needs = [counters._ACK(), counters._IP()]
    fmt = "{:.3f}"

    def compute(self, ack, ip):
        if ack == 0:
            return 0
        return 100. * ack / ip


class R_ICMP(AtomicStat):
    """
    Ratio of ICMP packets
    """

    description = "Ratio of ICMP packets"
    needs = [counters._ICMP(), counters._IP()]
    fmt = "{:.3f}"

    def compute(self, icmp, ip):
        if icmp == 0:
            return 0
        return 100. * icmp / ip


class AVG_PKT_BYTES(AtomicStat):
    """
    Average size of IP packets
    """

    description = "Average size of IP packets"
    needs = [counters._IP_BYTES(), counters._IP()]
    fmt = "{:.3f}"

    def compute(self, byte, ip):
        if byte == 0:
            return 0
        return 1. * byte / ip


class NB_IP_PKTS(AtomicStat):
    """
    Number of IP packets
    """

    description = "Number of IP packets"
    needs = [counters._IP()]
    fmt = "{:d}"

    def compute(self, ip):
        return ip


class SRC_DST_RATIO(AtomicStat):
    """
    Ratio (unique src addr) / (unique dst addr)
    """

    description = "Ratio (unique src addr) / (unique dst addr)"
    needs = [counters._UNIQUE_SRC_ADDR(), counters._UNIQUE_DST_ADDR()]
    fmt = "{:.3f}"

    def compute(self, nb_unique_srcaddr, nb_unique_dstaddr):
        if nb_unique_srcaddr == 0:
            return 0
        return 1. * nb_unique_srcaddr / nb_unique_dstaddr


class NB_IP_TO_IP_PKTS(AtomicStat):
    """
    Number of pkts between 2 IP
    """

    description = "Number of pkts between 2 IP"
    needs = []
    # fmt = "{:" + str(len(__name__)) + ".3f}"
    fmt = "{:d}"

    def __init__(self, ip_a, ip_b, **kwargs):
        super(NB_IP_TO_IP_PKTS, self).__init__()
        self.needs = [counters._IP_TO_IP(ip_a, ip_b)]
        self.name = "NB_{}_TO_{}_PKTS".format(ip_a, ip_b)

    def compute(self, ip_to_ip):
        return ip_to_ip


###############################################################################
###############################################################################
###############################################################################
# MISC
###############################################################################
###############################################################################
###############################################################################
AVAILABLE_STATS = utils.get_pure_source_classes(sys.modules[__name__])


def is_requiring_parameters(stat_class):
    if stat_class in AVAILABLE_STATS.values():
        return len(inspect.signature(stat_class).parameters) > 1


def stat_from_name(name, *extra):
    if name in AVAILABLE_STATS.keys():
        if is_requiring_parameters(AVAILABLE_STATS[name]):
            return AVAILABLE_STATS[name](*extra)
        else:
            return AVAILABLE_STATS[name]()
    else:
        raise ValueError("Unknown statistics")
