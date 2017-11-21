# Gossip protocol

Package contains tools for simulation of net based on gossip protocol.
You can see the behaviour in the session log or measure speed and net 
load externally.
See *Implementation section* for more details.

## Installation
You can get and install package with standart gotool:
```console
$ go get github.com/sokks/gossip
```
To run task used for graph construction (WARNING: it takes long time):
```console
$ cd `go env GOPATH`/src/github.com/sokks/gossip/task2/
$ sudo make task2
$ make draw
```
To run peformance data collection (*currently unavailable*):
```console
$ sudo su
# go install github.com/sokks/gossip/performance/
# performance <n_of_nodes> <base_port> <min_degree> <max_degree> <ttl> <session_log_dir>
```

## Dependencies
The package uses graph package for representation of the net (**gitlab.com/n-canter/graph**).

## Usage
There is API 

## Implementation
Transport protocol: **UDP**  
Network interface: **loopback**

*Algorithm:*  
Each node is running in seporate goroutine. A node has its own sender, reciever and internal processor.
A node launches an infinite loop for communication and is controled via channels.
```go
select {
case <-kill:
    // stop processing
case <-receive:
    // process message
case <-ticker:
    // send random message to random peer
    // send random ack to random peer
} 
```
**Load balance:** Non-blocking receive provides load balancing.  
**Flood prevention:** Node caches received messages to prevent double-sending.

### task2
Tests were made with following parameters:
- graph size = 50
- min degree = 5
- max degree = 7
- TTL = 100
Load of *lo* interface about 1000 packets/sec (500/500)

## Examples
There *will be* graphs representing performance on some metrics in the *performance* folder:
- size of net
- average number of links for one node
- probability of packages loss
- load of network interface
