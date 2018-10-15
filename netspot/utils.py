#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Wed Sep 19 15:44:53 2018

@author: asr
"""

import inspect
import logging
import os
import socket


def get_pure_source_classes(module):
    """
    Returns the classes defined within a module (neither the abstract nor the imported classes)
    """
    members = inspect.getmembers(
        module, predicate=lambda e: inspect.isclass(e) and (
            not inspect.isabstract(e)) and (
            inspect.getmodule(e).__name__ == module.__name__))
    return dict(members)


LOGGING_LEVEL = {logging.INFO: 'INFO',
                 logging.WARNING: 'WARNING',
                 logging.ERROR: 'ERROR',
                 logging.CRITICAL: 'CRITICAL',
                 logging.DEBUG: 'DEBUG'}


class ShellColor:
    """
    Just a basic object to print on shell with colors
    """
    ESCAPE_CHAR = "\x1B[{}m"

    color_map = {
        'black': 30,
        'red': 31,
        'green': 32,
        'yellow': 33,
        'blue': 34,
        'violet': 35,
        'cyan': 36,
        'gray': 37,
        'white': 97
    }
    END = "\x1B[0m"
    RED = "\x1B[31m"
    GRN = "\x1B[32m"
    YEL = "\x1B[33m"
    BLU = "\x1B[34m"
    VIO = "\x1B[35m"
    CYN = "\x1B[36m"
    GRA = "\x1B[37m"
    WHT = "\x1B[97m"
    BLA = "\x1B[30m"
    BOLD = "\x1B[1m"
    UNDERLINED = "\x1B[4m"
    ITALIC = "\x1B[3m"

    OK = '[' + "\x1B[92m" + 'OK' + "\x1B[0m" + ']'
    NO = '[' + "\x1B[91m" + 'NO' + "\x1B[0m" + ']'


def italic(text):
    if not isinstance(text, str):
        text = str(text)
    return ShellColor.ITALIC + text + ShellColor.END


def colorize(
        text,
        color,
        light=False,
        bold=False,
        underlined=False,
        background=False):
    """
    Style formatting function (especially for the color)
    """
    if isinstance(text, Exception):
        text = str(text)
    if color in ShellColor.color_map:
        code = ShellColor.color_map[color]
        if light:
            code += 60
        if background:
            code += 10
        start = ShellColor.ESCAPE_CHAR.format(code)
        if bold:
            start = ShellColor.BOLD + start
        if underlined:
            start = ShellColor.UNDERLINED + start
        return start + text + ShellColor.END


def _is_valid_ipv4_address(address):
    """
    Check if the string input corresponds to a valid IPv4 address
    """
    try:
        socket.inet_pton(socket.AF_INET, address)
    except AttributeError:  # no inet_pton here, sorry
        try:
            socket.inet_aton(address)
        except socket.error:
            return False
        return address.count('.') == 3
    except socket.error:  # not a valid address
        return False
    return True


def _is_valid_ipv6_address(address):
    """
    Check if the string input corresponds to a valid IPv6 address
    """
    try:
        socket.inet_pton(socket.AF_INET6, address)
    except socket.error:  # not a valid address
        return False
    return True


def print_error(msg, **kwargs):
    """
    Print in light red
    """
    print(colorize(msg, 'red', light=True), **kwargs)


def print_ok(msg, **kwargs):
    """
    Print in light green
    """
    print(colorize(msg, 'green', light=True), **kwargs)


def print_warning(msg, **kwargs):
    """
    Print in yellow
    """
    print(colorize(msg, 'yellow'), **kwargs)


def is_valid_address(address):
    """
    Check if the string input corresponds to a valid IPv4/IPv6 address
    """
    return _is_valid_ipv4_address(address) | _is_valid_ipv6_address(address)


def is_iterable(obj):
    """
    Check if an object is iterable or not
    """
    return hasattr(obj, '__getitem__')


def to_int(x, custom_error_msg=None):
    """
    Try to convert to integer and raise an exception (with a custom message) if it fails
    """
    try:
        return int(x)
    except ValueError as e:
        if custom_error_msg is None:
            raise ValueError(e)
        else:
            raise ValueError(custom_error_msg)


def is_file_ok(file):
    """
    Check if a file is valid (= we can read or create it)
    """
    if file:
        path = os.path.abspath(file)
        basename = os.path.basename(path)
        folder = os.path.dirname(path)
        return os.path.isdir(folder) and basename != ''
    return False
