import pandas as pd;
import matplotlib as mpl
import matplotlib.pyplot as plt
import matplotlib.style as style
import seaborn as sns
from matplotlib.colors import ListedColormap

import argparse

def parse_args():
    parser = argparse.ArgumentParser(description='Generate report')
    parser.add_argument('-source', action="store", dest="source", default="input")
    return parser.parse_args()

args = parse_args()




sns.set_style('darkgrid', {'legend.frameon':True})

#style.use('fivethirtyeight')
mpl.rcParams["axes.formatter.useoffset"] = False
# get rid of scientific notation
plt.ticklabel_format(style='plain', axis='y')


df = pd.read_table(args.source)
# skip first row
df = df.drop(df.index[0])



fig, ax = plt.subplots(1)
fig.suptitle("Thoughput")
ax.set_ylabel("Hz")
ax.set_xlabel("Total transactions")

ax.plot(df["TxTotal"],df["Hz"])
ax.set_ylim(bottom=0)
fig.savefig("throughput.png")
plt.close(fig)
print("Saved throughput.png")




fig, ax = plt.subplots(1, sharex = True)
fig.suptitle("Latency over time")
ax.set_ylabel("Latency ms")
ax.set_xlabel("Transactions performed")

percentiles = ["P50", "P90", "P99", "P999", "100"]

p_count = len(percentiles)
cmap = ListedColormap(sns.color_palette("YlGnBu_r", len(percentiles)))


prev = 0

for idx, name in enumerate(percentiles):
    line = df[name]
    alpha = 1.0 ;
    label = 'Percentile {0}'.format(name)
    ax.fill_between(df['TxTotal'], prev, line, alpha=alpha, color=cmap(idx), label=label, interpolate=True)
    prev = line

ax.legend(loc='upper left', ncol=1, fancybox=True,framealpha=0.5, facecolor="white")



fig.savefig("latency.png")
plt.close(fig)

print("Saved latency.png")
