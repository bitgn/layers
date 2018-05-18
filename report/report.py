import pandas as pd;
import matplotlib as mpl
import matplotlib.pyplot as plt
import matplotlib.style as style
import seaborn as sns
from matplotlib.colors import ListedColormap
from pathlib import Path
import argparse
import json


def parse_args():
    parser = argparse.ArgumentParser(description='Generate report')
    parser.add_argument('-source', action="store", dest="source", default="input")
    return parser.parse_args()

args = parse_args()


folder = Path(args.source)

meta = folder.joinpath("meta.json")



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
setup =  """FoundationDB: {0}x {1}, {2} {3}; tester: 1x {4} - at {5} Hz
{6}""".format(meta_count, meta_vm, status_engine, status_redundancy, meta_tester,
           bench_hz,
           bench_setup)



tsv_file = folder.joinpath(meta["main_tsv"])

sns.set_style('darkgrid', {'legend.frameon':True})

#style.use('fivethirtyeight')
mpl.rcParams["axes.formatter.useoffset"] = False
# get rid of scientific notation
plt.ticklabel_format(style='plain', axis='y')


df = pd.read_table(tsv_file)
# skip first row
df = df.drop(df.index[0])



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

fig.savefig("summary.png")
plt.close(fig)

print("Saved summary.png")



fig, axs = plt.subplots(2, sharex = True)
ax=axs[0]

ax.plot(df["Seconds"],df["Move"])
ax.set_ylim(bottom=0)


ax=axs[1]

ax.plot(df["Seconds"],df["KVTotal"])
ax.plot(df["Seconds"],df["Disk"])
ax.set_ylim(bottom=0)

fig.savefig("data.png")
plt.close(fig)

print("Saved data.png")


fig, axs = plt.subplots(2, sharex = True)
ax=axs[0]

fig.suptitle(experiment)
fig.subplots_adjust(top=0.85)
title = ax.set_title(setup)

ax.plot(df["Seconds"],df["Conflicted"])
ax.set_ylim(bottom=0)

ax.set_ylabel("Conflicted Tx Hz")


ax = axs[1]
ax.set_xlabel("Seconds")


fig.savefig("tx.png")
plt.close(fig)

print("Saved tx.png")
