import argparse
import copy
import os
import sys
from itertools import chain, islice

import matplotlib.pyplot as plt
import numpy as np

if sys.version_info >= (3, 0):
    from itertools import zip_longest as izip
else:
    from itertools import izip as izip


data_file = "test_loss.data"
plot_file = "test_loss.png"
boxplot_file = "test_loss_box.png"

parser = argparse.ArgumentParser()
parser.add_argument(
    '--path', help='path to directory where to store data', required=True)
args = parser.parse_args()
data_dir = args.path

datapath = os.path.join(data_dir, data_file)
plotpath = os.path.join(data_dir, plot_file)
boxplotpath = os.path.join(data_dir, boxplot_file)
probs, probs1, data_raw, boxes, avgs = [], [], [], [], []
n_experiments = None

with open(datapath) as filein, open(datapath) as filein2:
    n_experiments = int(filein.readline())
    for line1, line2 in izip(islice(filein, 0, None, 2), islice(filein2, 2, None, 2)):
        probs1.append(float(line1))
        np_ar = np.fromiter((t for t in line2.split()), dtype=np.int)
        boxes.append(sorted(np_ar))
        data_raw.append(np_ar[1:-1])
        avgs.append(np.average(np_ar))


x = np.fromiter(chain.from_iterable(
    ([t] * (n_experiments - 2) for t in probs1)), dtype=np.float)
y = np.concatenate(data_raw)
x1 = probs1
y1 = avgs

plt.clf()
plt.axis([-0.1, max(x) + 0.1, 0, max(y) + 20])
plt.scatter(x, y, s=30, c='green', marker='o', alpha=0.6)
plt.savefig(plotpath)
plt.plot(x1, y1, 'ro-')
plt.savefig(plotpath)

plt.clf()
fig = plt.figure()
ax = fig.add_subplot(111)
bp = ax.boxplot(boxes, patch_artist=True)

# <stackoverflow>
for box in bp['boxes']:
    box.set(color='#7570b3', linewidth=2)
    box.set(facecolor='#1b9e77')

for whisker in bp['whiskers']:
    whisker.set(color='#7570b3', linewidth=2)

for cap in bp['caps']:
    cap.set(color='#7570b3', linewidth=2)

for median in bp['medians']:
    median.set(color='#b2df8a', linewidth=2)

for flier in bp['fliers']:
    flier.set(marker='o', color='#e7298a', alpha=0.5)

ax.set_xticklabels(x1)
plt.savefig(boxplotpath, bbox_inches='tight')

# </stackoverflow>
