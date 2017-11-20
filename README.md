# Gossip protocol

Package contains tools for simulation of net based on gossip protocol.
You can see the поведение in the session log or measure speed and net 
load externally.
See Implementation section for more details.

Currently loopback interface is used. It's further work of implementation 
distributed nodes.

### Installation
You can get and install package with standart gotool:
```Markdown
# go get github.com/sokks/gossip
```
To run task used for graph construction (root mode):
```Markdown
# go install github.com/sokks/gossip/main/
# main <sessoin_log_dir>
or
# $GOPATH/bin/main <session_log_dir>
```
To run peformance data collection:
```Markdown
# go install github.com/sokks/gossip/performance/
# performance <n_of_nodes> <base_port> <min_degree> <max_degree> <ttl> <session_log_dir>
```

### Dependencies
The package uses graph package for representation of the net (**gitlab.com/n-canter/graph**).

### Usage
There is API 

### Implementation
UDP
Algorithm used:
> 
Flood prevention
Node reciever sender processor
message queue

### Examples
There are graphs representing performance on some metrics in the *performance* folder:
- size of net
- average number of links for one node
- probability of packages loss
- load of network interface
