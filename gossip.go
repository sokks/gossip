package gossip

import (
	"strconv"
	"net"
	"gitlab.com/n-canter/graph"
	"fmt"
	"time"
	"log"
	"os"
)

const (
	NET_SIZE = 10
	BASE_PORT = 9080
	TTL = 10
)

var (
	logger *log.Logger
	logfile *os.File
)


// GossipNode represents one gossip net peer.
// It works with UDP connection using internal sender and receiver.
type GossipNode struct {
	id 			int
	port 		int
	udpConn 	*net.UDPConn
	receiver 	*Receiver
	sender 		*Sender
	processor 	*nodeProcessor
}

// NewGossipNode constracts new GossipNode based on its graph place.
func NewGossipNode (id int, port int, neighs []graph.Node) *GossipNode {
	return &GossipNode{
		id: 		id,
		port: 		port,
		udpConn: 	nil,
		receiver: 	nil,
		sender: 	nil,
		processor: 	newNodeProcessor(id, neighs),
	}
}

// Bind creates a socket on the loopback interface with this node's port
// and assosiates sender and receiver with this UDP connection.
func (gn *GossipNode) Bind() {
	laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(gn.port)))
	if err != nil {
		logger.Println(time.Now().String(), "NODE", gn.id, "ERROR: cannot resolve address")
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		logger.Println(time.Now().String(), "NODE", gn.id, "ERROR: cannot bind port")
	}
	gn.udpConn = conn
	logger.Printf("[NODE %d] port binded", gn.id)
	gn.receiver = NewReceiver(conn)
	gn.sender = NewSender(conn)
}

// Unbind closes socket
func (gn *GossipNode) Unbind() {
	gn.udpConn.Close()
	logger.Printf("[NODE %d] port unbinded", gn.id)
}

func (gn *GossipNode) putNewRumour(msg Message, netSize int) (exists bool) {
	// give a command to processor to put message in the queue and start tracking it
	return gn.processor.initNewMessage(msg, netSize) 
}

// Process sends messages to random peers every interval and processes incoming messages
func (gn *GossipNode) Process(kill chan struct{}, interval time.Duration) {
	logger.Printf("[NODE %d] started processing", gn.id)
	gn.Bind()
	defer gn.Unbind()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	gn.receiver.Start()
	defer gn.receiver.Stop()
	gn.sender.Start()
	defer gn.sender.Stop()
	counter := 0
	for {
		select {
		case <-kill:  // got stop signal
			return
		case msg := <-gn.receiver.C: // got some message from receiver
			logger.Printf("[NODE %d] message received %s", gn.id, msg.String())
			gn.processor.processMsg(msg)
		case <-ticker.C:  // time for new round
			counter++
			msg, addr, empty := gn.processor.getRandomMsg()
			if !empty {
				logger.Printf("[NODE %d] sending to address %s message %s", gn.id, addr, msg)
				gn.sender.C <- senderPack{msg, addr} // sending task for node's sender
			}
			msg, addr, empty = gn.processor.getRandomAck()
			if !empty {
				logger.Printf("[NODE %d] sending to address %s ack %s", gn.id, addr, msg)
				gn.sender.C <- senderPack{msg, addr}
			}
		}
	}
}

// GossipNet represents whole net. It consists of several nodes 
// and can be constructed of graph(TODO) or randomly.
type GossipNet struct {
	size  int
	nodes []*GossipNode
	kill  chan struct{}
}

func initLogger(logDirectoryPath string) *log.Logger {
	var err error
	logfile, _ = os.OpenFile(logDirectoryPath + "session_" + time.Now().Format("20060102150405") + ".log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {fmt.Println("Couldn't open file")}
	return log.New(logfile, "TRACE:", log.Ltime)
}

func closeLogger() {
	logger.Println("Closing log")
	logfile.Close()
}

// InitNet generates random net of size n. It uses graph package to create graph.
func InitNet(n int) *GossipNet {
	g := graph.Generate(n, 1, 5, BASE_PORT)
	GNs := make([]*GossipNode, 0, n)
	for i := 0; i < n; i++ {
		node, _ := g.GetNode(i)
		nodeId, _ := strconv.Atoi(node.String())
		nodePort := node.Port()
		neighs, _ := g.Neighbors(i)
		gn := NewGossipNode(nodeId, nodePort, neighs)
		GNs = append(GNs, gn)
	}
	return &GossipNet{
		size: n,
		nodes: GNs,
		kill: make(chan struct{}, n),
	}
}

// InitNetFromGraph generates net from input graph of package graph.
// NOTE: It doesn't check the graph correctness.
func InitNetFromGraph(g graph.Graph) *GossipNet {
	n := len(g)
	GNs := make([]*GossipNode, 0, n)
	for i := 0; i < n; i++ {
		node, _ := g.GetNode(i)
		nodeId, _ := strconv.Atoi(node.String())
		nodePort := node.Port()
		neighs, _ := g.Neighbors(i)
		gn := NewGossipNode(nodeId, nodePort, neighs)
		GNs = append(GNs, gn)
	}
	return &GossipNet{
		size: n,
		nodes: GNs,
		kill: make(chan struct{}, n),
	}
}


// TODO: interval as parameter of the net
// TODO: logger filepath as parameter of Start()

// Start lanches the gossip simulation. Also it inits the session logger.
// Each node is launched in the sepotare goroutine.
func (GN *GossipNet) Start() {
	//logger = initLogger("D:\\Documents\\sem7\\clouds_hpc\\task2_gossip\\")
	logger = initLogger("")
	logger.Println(time.Now().String(), " Start")
	for i:= 0; i < GN.size; i++ {
		go GN.nodes[i].Process(GN.kill, 5 * time.Second)
	}
	time.Sleep(time.Second)
}

// Stop sends stop signals to nodes and closes the session logger.
func (GN *GossipNet) Stop() {
	for i := 0; i <= GN.size; i++ {
		GN.kill <- struct{}{}
		time.Sleep(50 * time.Millisecond)
	}
	closeLogger()
}

type errorString struct {
    s string
}

func (e *errorString) Error() string {
    return e.s
}

// MakeRumour inits node with id id to generate new message and start tracking it.
func (GN *GossipNet) MakeRumour(id int, msg Message) error {
	exists := GN.nodes[id].putNewRumour(msg, GN.size)
	if exists {
		return &errorString{"such message ID has been sent already"}
	}
	return nil
}
