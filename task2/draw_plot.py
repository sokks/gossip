import os
import sys

import matplotlib.pyplot as plt
import numpy as np

if len(sys.argv) < 2:
    print("Not enough arguments: data directory needed")
data_dir = sys.argv[1]
data_file = "test_loss.data"
plot_file = "test_loss.png"
datapath = os.path.join(data_dir, data_file)
plotpath = os.path.join(data_dir, plot_file)
print(datapath)
probs = []
data_raw = []
with open(datapath) as filein:
    nn = filein.readline()
    n_experiments = int(nn)
    print(n_experiments)
    line = filein.readline()
    while len(line) > 0:
        prob = float(line)
        probs.append(prob)
        line = filein.readline()
        res = line.split()
        data_raw.append(res)
        probs += [prob] * (n_experiments - 1)
        line = filein.readline()
print(probs)
print(data_raw)
x = np.array(probs)
y = np.concatenate(data_raw)
#y = np.array(data_raw)
print(x, y)
max_x = max(probs)
max_y = max(y)
print(max(probs), max(y))

plt.scatter(x, y, s=30, c='green', marker='o', alpha=0.6)
plt.axis([-0.1, max_x + 0.1, 0, 400])
plt.savefig(plotpath)
