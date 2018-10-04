#!/usr/bin/python3
# -*- coding: utf-8 -*-
"""
@author: asr
"""

from setuptools import setup, find_packages

setup(name='netspot',
      version='1.0',
      description='A basic IDS with statistical learning',
      author='Alban Siffer',
      author_email='alban.siffer@irisa.fr',
      license='GPL-3',
      packages=find_packages(),
      install_requires=['pylibspot', 'scapy', 'pandas', 'netifaces', 'configparser'],
      entry_points={
        'console_scripts': ['netspot=netspot.run:main'],
      },
      data_files=[("etc/netspot", ["netspot.ini"])],
      zip_safe=False)
