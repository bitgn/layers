import pandas as pd;
import matplotlib as mpl
import matplotlib.pyplot as plt
import matplotlib.style as style
import seaborn as sns
from matplotlib.colors import ListedColormap
from pathlib import Path
import argparse
import json

import htmlgen as html


def parse_args():
    parser = argparse.ArgumentParser(description='Generate report')
    parser.add_argument('-source', action="store", dest="source", default="input")
    return parser.parse_args()

args = parse_args()


folder = Path(args.source)

def rel(name):
    return str(folder.joinpath(name))

meta = rel("meta.json")



with open(folder.joinpath("meta.json")) as f:
    meta = json.load(f)

with open(folder.joinpath(meta["status_file"])) as f:
    status = json.load(f)



status_engine = status["cluster"]["configuration"]["storage_engine"]
status_redundancy = status["cluster"]["configuration"]["redundancy"]["factor"]
status_processes = len(status["cluster"]["processes"])

meta_vm = meta["cluster"]["fdb_type"]
meta_count = meta["cluster"]["fdb_count"]
meta_tester = meta["cluster"]["tester_type"]

experiment = meta["bench-name"]
bench_setup = meta["bench-setup"]
bench_hz = meta["bench-hz"]

if bench_hz < 0:
    speed = "{0} actors".format(-bench_hz)
    bench_concurrency = speed
    bench_throughput = "adaptive"
else:
    speed = "{0} Hz".format(bench_hz)
    bench_concurrency = "adaptive"
    bench_throughput = speed


setup =  """FDB: {0}*{1} ({7} procs), {2} {3}; load: 1*{4} - {5}
{6}""".format(meta_count, meta_vm, status_engine, status_redundancy, meta_tester,
              speed,
              bench_setup,
              status_processes
)



tsv_file = folder.joinpath(meta["main_tsv"])

sns.set_style('darkgrid', {'legend.frameon':True})

#style.use('fivethirtyeight')
mpl.rcParams["axes.formatter.useoffset"] = False
# get rid of scientific notation
plt.ticklabel_format(style='plain', axis='y')


def legend(ax):
    ax.legend(loc='upper left', ncol=1, fancybox=True,framealpha=0.5, facecolor="white")


df = pd.read_table(tsv_file)
# skip first row
df = df.drop(df.index[0])

print("Median Hz: {0}".format(df["Hz"].median()))


fig, axs = plt.subplots(2, sharex = True)
ax=axs[0]
fig.suptitle(experiment)
fig.subplots_adjust(top=0.85)
title = ax.set_title(setup)
ax.set_ylabel("Transactions per sec")
#ax.set_xlabel("Total transactions")

ax.plot(df["Seconds"],df["Hz"])
ax.set_ylim(bottom=0)


# fig.savefig("throughput.png")
#plt.close(fig)
#print("Saved throughput.png")




#fig, ax = plt.subplots(1, sharex = True)
ax=axs[1]
# fig.suptitle("Latency: " + experiment)

# plt.title(setup)
ax.set_ylabel("Latency ms")

percentiles = ["P50", "P90", "P99", "P999", "P100"]

p_count = len(percentiles)
cmap = ListedColormap(sns.color_palette("YlGnBu_r", len(percentiles)))


prev = 0

for idx, name in enumerate(percentiles):
    line = df[name]
    alpha = 1.0 ;
    label = 'Percentile {0}'.format(name)
    ax.fill_between(df['Seconds'], prev, line, alpha=alpha, color=cmap(idx), label=label, interpolate=True)
    prev = line

ax.legend(loc='upper left', ncol=1, fancybox=True,framealpha=0.5, facecolor="white")


ax.set_xlabel("Seconds")

fig.savefig(rel("summary.png"))
plt.close(fig)



### DATA CHART
fig, axs = plt.subplots(3, sharex = True)
fig.suptitle("Data")

ax=axs[0]
ax.plot(df["Seconds"],df["KVTotal"], label="Sum of key-value sizes")
ax.set_ylabel("MB")

ax.plot(df["Seconds"],df["Disk"], label="Disk space used")
ax.set_ylim(bottom=0)
legend(ax)

ax=axs[1]
ax.plot(df["Seconds"],df["Move"], label="Moving data")
ax.set_ylim(bottom=0)
ax.set_ylabel("MB")
legend(ax)


ax=axs[2]
ax.plot(df["Seconds"],df["Partitions"], label="Partitions")
ax.set_ylim(bottom=0)
legend(ax)

ax.set_xlabel("Seconds")

fig.savefig(rel("data.png"))
plt.close(fig)



fig, axs = plt.subplots(2, sharex = True)

fig.suptitle("Workload")
ax=axs[0]
ax.plot(df["Seconds"],df["Conflicted"], label="Transaction conflicts")
legend(ax)
ax.set_ylim(bottom=0)
ax.set_ylabel("Conflicted Tx Hz")


ax = axs[1]
ax.plot(df["Seconds"], df["Queue1"], label="Client: waiting to launch")
ax.plot(df["Seconds"], df["Queue2"], label="Client: waiting for response")
legend(ax)

ax.set_xlabel("Seconds")


fig.savefig(rel("tx.png"))
plt.close(fig)


def html_list(*items):
    html = "<ul>"
    for (k, v) in items:
        html += "<li>{0}: <em>{1}</em></li>".format(k,v)
    return html + "</ul>"

build = meta["build"]

if not build:
    build = "dev"

config_list = html_list(
    ("<strong>Benchmark</strong>", experiment),
    ("<strong>FoundationDB Cluster</strong>", "{0}x <code>{1}</code> nodes ({2} processes total)".format(meta_count, meta_vm, status_processes)),
    ("<strong>Storage</strong>", "{0} {1}".format(status_engine, status_redundancy)),
    ("<strong>Throughput</strong> per tester", bench_throughput),
    ("<strong>Concurrency</strong> per tester", bench_concurrency),
    ("Benchmark config", bench_setup),
    ("Arguments", "<code>" + " ".join(meta["args"]) + "</code>"),
    ("Build", "<code>" + build + "</code>"),
    ("Test VMs", "{0}x <code>{1}</code>".format(1, meta_tester)),
    ("Time", meta["time"])
)
        


with open(rel("index.html"), "w") as f:


    
    text = meta["bench-text"]

    if text:
        lines = text.split("\n")
        for l in lines:
            config_list += "<p>" + l + "</p>"
    
    b = html.simple_sect(
        "Setup",
        config_list,
    )

        

    b+= html.simple_sect(
        "Charts",
        html.img("summary.png"),
        html.img("tx.png"),
        html.img("data.png"),

        )

    b+= '<p class="text-right">by <a href="https://abdullin.com">Rinat Abdullin</a></p>'

    
    index = html.page(
        experiment + " | FoundationDB Benchmark",
        html.nav_bar("FoundationDB Benchmark"),
        b)
    f.write(index)



