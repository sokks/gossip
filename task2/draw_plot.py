import os
import sys

import matplotlib.pyplot as plt
import numpy as np

if len(sys.argv) < 2:
    print("Not enough arguments: data directory needed")
data_dir = sys.argv[1]
data_file = "test_loss.data"
plot_file = "test_loss.png"
boxplot_file = "test_loss_box.png"
datapath = os.path.join(data_dir, data_file)
plotpath = os.path.join(data_dir, plot_file)
boxplotpath = os.path.join(data_dir, boxplot_file)
probs = []
probs1 = []
data_raw = []
boxes = []
with open(datapath) as filein:
    nn = filein.readline()
    n_experiments = int(nn)
    line = filein.readline()
    while line:
        prob = float(line)
        probs += [prob] * (n_experiments - 2)
        probs1.append(prob)
        row = sorted(list(map(int, filein.readline().split())))
        data_raw.append(row[1:len(row) - 1])
        boxes.append(np.array(row))
        line = filein.readline()
x = np.array(probs)
y = np.concatenate(data_raw)
avgs = [sum(l)/len(l) for l in data_raw]
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
## change outline color, fill color and linewidth of the boxes
for box in bp['boxes']:
    # change outline color
    box.set( color='#7570b3', linewidth=2)
    # change fill color
    box.set( facecolor = '#1b9e77' )

## change color and linewidth of the whiskers
for whisker in bp['whiskers']:
    whisker.set(color='#7570b3', linewidth=2)

## change color and linewidth of the caps
for cap in bp['caps']:
    cap.set(color='#7570b3', linewidth=2)

## change color and linewidth of the medians
for median in bp['medians']:
    median.set(color='#b2df8a', linewidth=2)

## change the style of fliers and their fill
for flier in bp['fliers']:
    flier.set(marker='o', color='#e7298a', alpha=0.5)

ax.set_xticklabels(['0.0', '0.1', '0.2', '0.3', '0.4', '0.5'])
plt.savefig(boxplotpath, bbox_inches='tight')
