#!/bin/env python3

import pygal
from pygal.style import RedBlueStyle
from pygal.style import Style
import os
import pandas as pd

asset_folder = "/home/asr/go/src/netspot/sources/assets/"

colors = ["#5e81ac", "#81a1c1", "#8fbcbb", "#4c566a"]

custom_style = RedBlueStyle()
custom_style.colors = colors
custom_style.background = 'transparent'
custom_style.title_font_size = 18
custom_style.major_label_font_size = 18
custom_style.label_font_size = custom_style.major_label_font_size
custom_style.legend_font_size = custom_style.major_label_font_size

for attr in ["font_family",
             "label_font_family",
             "major_label_font_family",
             "value_font_family",
             "value_label_font_family",
             "tooltip_font_family",
             "title_font_family",
             "legend_font_family"]:
    setattr(custom_style, attr, '"Roboto"')


desktop = {
    "netspot v2.0a": 1000000,
    "netspot v1.3": 500000,
    "suricata": 400000,
    "kitsune": 40000,
}

rpi = {
    "netspot v2.0a": 100000,
    "netspot v1.3": 50000,
    "suricata": None,
    "kitsune": 5000,
}

# DESKTOP
bar = pygal.Bar(max_scale=5,
                legend_at_bottom=True,
                legend_at_bottom_columns=4,
                y_title="Packets per second",
                style=custom_style)
bar.title = None
for k, v in desktop.items():
    bar.add(k, v)

svg = bar.render()
svg = svg.replace(b"33px", b"1em")
with open(os.path.join(asset_folder, "perf-desktop.svg"), 'wb') as w:
    w.write(svg)


# RPI
bar = pygal.Bar(max_scale=5,
                legend_at_bottom=True,
                legend_at_bottom_columns=4,
                y_title="Packets per second",
                style=custom_style)
bar.title = None
for k, v in rpi.items():
    bar.add(k, v)

svg = bar.render()
svg = svg.replace(b"33px", b"1em")
with open(os.path.join(asset_folder, "perf-rpi.svg"), 'wb') as w:
    w.write(svg)


# GOMAXPROCS
folder = "/home/asr/go/src/netspot/.dev/perfs"

custom_style.colors = ["#457b9d"]
box_plot = pygal.Box(max_scale=10,
                     style=custom_style,
                     y_title="Packets per second",
                     x_title="GOMAXPROCS",
                     box_mode="stdev",
                     show_legend=False)
box_plot.x_labels = map(str, range(1, 7))


R = pd.DataFrame(index=range(1, 7, 1), columns=["mean", "std"])
for f in filter(lambda s: s.endswith('.json'), sorted(os.listdir(folder))):
    nb_procs = f.replace("netspot.", "").strip(".json")
    index = int(nb_procs)
    P = pd.read_json(os.path.join(folder, f), lines=True)
    R.loc[index, "mean"] = P["PERF"].mean()
    R.loc[index, "std"] = P["PERF"].std()

    data = P["PERF"].dropna().map(int).tolist()
    box_plot.add(nb_procs, data)

with open(os.path.join(asset_folder, "perf-procs.svg"), 'wb') as w:
    w.write(box_plot.render())
