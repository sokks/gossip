package gossip

import (
	"strings"
	"strconv"
	"net"
	"gitlab.com/n-canter/graph"
	"fmt"
	"encoding/json"
	"time"
	"log"
	"os"
	"math/rand"
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


// Message ...
type Message struct {
	ID int			`json:"id"`
	MsgType string	`json:"type"`
	Sender int		`json:"sender"`
	Origin int		`json:"origin"`
	Data string		`json:"data"`
}

func newMessage(id int, msgType string, sender int, origin int, data string) Message {
	return Message{ id, msgType, sender, origin, data}	
}

func (m Message) String() string {
	return fmt.Sprintf("{ ID: %d\n MsgType: %s\n Sender: %d\n Origin: %d\n Data: %s }",  
		m.ID, m.MsgType, m.Sender, m.Origin, m.Data)
}

type preparedMessage struct {
	msg Message
	distributionList []int
	ttl int
}

func newPreparedMessage(msg Message, list []int, ttl int) (*preparedMessage) {
	return &preparedMessage{msg, list, ttl}
}

type messageQueue struct {
	q []*preparedMessage
}

func newMessageQueue() *messageQueue {
	return &messageQueue{q: make([]*preparedMessage, 0, 100)} // ??
}

func (q *messageQueue) String() string {
	intSliceToString := func(values []int) string {
		valuesText := []string{}
		for _, val := range values {
			text := strconv.Itoa(val)
			valuesText = append(valuesText, text)
		}
		return "[ " + strings.Join(valuesText, " ") + " ]"
	}

	res := ""
	for _, val := range q.q {
		res += strconv.Itoa(val.msg.ID) + " "
		res += intSliceToString(val.distributionList)
		res += "\n"
	}
	return res
}

func (q *messageQueue) putMessage(msg Message, recipients []int) {
	fmt.Println("PUTTING TO", q, "slice:", q.q)
	q.q = append(q.q, newPreparedMessage(msg, recipients, TTL))
	fmt.Println("PUT TO", q, "slice:", q.q)
}

func (q *messageQueue) getMessage() (msg Message, id int, empty bool) {
	fmt.Println("GETTING FROM", q, "len =", len(q.q))
	if len(q.q) == 0 {
		return Message{}, 0, true
	}
	r := int(rand.Int31n(int32(len(q.q))))
	message := q.q[r]
	t := int(rand.Int31n(int32(len((*message).distributionList))))
	recipient := (*message).distributionList[t]
	fmt.Println("TTL", q.q[r].ttl)
	q.q[r].ttl--
	fmt.Println("TTL", q.q[r].ttl)
	fmt.Println(q.q)
	if q.q[r].ttl == 0 {
		q.q = append(q.q[:r], q.q[r + 1:]...)
	}
	fmt.Println(q.q)
	fmt.Println("message to send: ", message.msg)
	return message.msg, recipient, false
}

type nodeProcessor struct {
	myID 		int
	neighbours 	map[int]*net.UDPAddr
	msgIDs 		[]int
	ackIDs 		map[int][]int
	msgQueue 	*messageQueue
	ackQueue 	*messageQueue
	acks 		map[int][]bool
}

func newNodeProcessor(id int, neighs []graph.Node) *nodeProcessor {
	getAddr := func(port int) (*net.UDPAddr) {
		laddr, _ := net.ResolveUDPAddr("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		return laddr
	}
	makeNeighMap := func() (map[int]*net.UDPAddr) {
		m := make(map[int]*net.UDPAddr)
		for _, node := range neighs {
			nid, _ := strconv.Atoi(node.String())
			m[nid] = getAddr(node.Port())
		}
		return m
	}
	
	return &nodeProcessor{
		myID: 		id,
		neighbours: makeNeighMap(),
		msgIDs: 	make([]int, 0, 10),
		ackIDs:		make(map[int][]int),
		msgQueue:	newMessageQueue(),
		ackQueue:	newMessageQueue(),
		acks:		make(map[int][]bool),
	}
}

func (p *nodeProcessor) initNewMessage(msg Message, netSize int) (exists bool) {
	existingID := func(msgId int) bool {
		for _, val := range p.msgIDs {
			if val == msgId {
				return true
			}
		}
		return false
	}

	getDestList := func() ([]int) {
		res := make([]int, 0, len(p.neighbours))
		for key := range p.neighbours {
			res = append(res, key)
		}
		return res
	}

	msgId := msg.ID
	if !existingID(msgId) {
		p.acks[msgId] = make([]bool, netSize)
		p.acks[msgId][p.myID] = true
		p.msgIDs = append(p.msgIDs, msgId)
		p.msgQueue.putMessage(msg, getDestList())
		return false
	}
	return true
}

func (p *nodeProcessor) processMsg(msg Message) {
	alreadyReceivedMsg := func(id int) (bool) {
		for _, val := range p.msgIDs {
			if val == id {
				return true
			}
		}
		return false
	}

	alreadyReceivedAck := func(msgId, nodeId int) (bool) {
		hasKey := false
		for key := range p.ackIDs {
			if key == msgId {
				hasKey = true
				break
			}
		}
		if hasKey {
			for _, val := range p.ackIDs[msgId] {
				if val == nodeId {
					return true
				}
			}
		}
		return false
	}

	memorizeMsgID := func(id int) {
		p.msgIDs = append(p.msgIDs, id)
	}

	memorizeAckID := func(msgId, nodeId int) {
		hasKey := false
		for key := range p.ackIDs {
			if key == msgId {
				hasKey = true
				break
			}
		}
		if hasKey {
			p.ackIDs[msgId] = make([]int, 1, 10)
			p.ackIDs[msgId][0] = nodeId
		} else {
			p.ackIDs[msgId] = append(p.ackIDs[msgId], nodeId)
		}
	}

	const (
		ALL = 0
		EXCEPTSENDER = 1
	)

	getDestList := func(mode int) ([]int) {
		res := make([]int, 0, len(p.neighbours))
		if mode == 0 {
			for key := range p.neighbours {
				res = append(res, key)
			}
		} else {
			for key := range p.neighbours {
				if key != msg.Sender {
					res = append(res, key)
				}
			}
		}
		return res
	}

	initedByMe := func(msgId int) bool {
		for key := range p.acks {
			if key == msgId {
				return true
			}
		}
		return false
	}

	writeAck := func(msgId, nodeId int) {
		p.acks[msgId][nodeId] = true
	}

	ackedByAll := func(msgId int) bool {
		for _, val := range p.acks[msgId] {
			if !val {
				return false
			}
		}
		
		return true
	}

	boolSliceToString := func(values []bool) string {
		valuesText := []string{}
		for _, val := range values {
			text := strconv.FormatBool(val)
			valuesText = append(valuesText, text)
		}
		return "[ " + strings.Join(valuesText, " ") + " ]"
	}

	if msg.MsgType == "multicast" {
		if !alreadyReceivedMsg(msg.ID) {
			memorizeMsgID(msg.ID)
			p.msgQueue.putMessage(Message{msg.ID, "multicast", p.myID, msg.Origin, msg.Data}, getDestList(EXCEPTSENDER))
			p.ackQueue.putMessage(Message{msg.ID, "notification", p.myID, p.myID, "ack"}, getDestList(ALL))
		}
	} else { // msg.MsgType == "notification" {
		if !alreadyReceivedAck(msg.ID, msg.Origin) {
			memorizeAckID(msg.ID, msg.Origin)
			if initedByMe(msg.ID) {
				logger.Printf("[NODE %d] [MESSAGE %d ACKED BY NODE %d] acks for message: %s \n", p.myID, msg.ID, msg.Origin, boolSliceToString(p.acks[msg.ID]))
				writeAck(msg.ID, msg.Origin)
				if ackedByAll(msg.ID) {
					logger.Printf("[NODE %d] [MESSAGE %d ACKED BY ALL NODES] time passed TODO\n", p.myID, msg.ID)
				}
			}
			p.msgQueue.putMessage(Message{msg.ID, "notification", p.myID, msg.Origin, msg.Data}, getDestList(EXCEPTSENDER))
		}
	}
}

func (p *nodeProcessor) getRandomMsg() (Message, *net.UDPAddr, bool) {
	getAddr := func(id int) *net.UDPAddr {
		return p.neighbours[id]
	}
	
	msg, nodeId, empty := p.msgQueue.getMessage()
	if empty {
		return Message{}, nil, true
	} else {
		return msg, getAddr(nodeId), false
	}
}

func (p *nodeProcessor) getRandomAck() (Message, *net.UDPAddr, bool) {
	getAddr := func(id int) *net.UDPAddr {
		return p.neighbours[id]
	}
	
	msg, nodeId, empty := p.ackQueue.getMessage()
	if empty {
		return Message{}, nil, true
	} else {
		return msg, getAddr(nodeId), false
	}
}

type Reciever struct {
	C chan Message
	kill chan struct{}
	udpConn *net.UDPConn
	buffer []byte
}

func NewReciever(udpConn *net.UDPConn) *Reciever {
	rcvr := &Reciever{make(chan Message, 100), make(chan struct{}), udpConn, make([]byte, 1024)} 
	return rcvr
}

func (r *Reciever) startReciever() {
	//timeout := time.Duration(100 * time.Millisecond)
	for {
		select {
		case <-r.kill:
			return
		default:
			msg := Message{}
			//r.udpConn.SetReadDeadline(time.Now().Add(timeout))
			n, _, err := r.udpConn.ReadFromUDP(r.buffer)
			if err != nil {
				if e, ok := err.(net.Error); !ok || !e.Timeout() {
					panic(err)
				} else {
					fmt.Println("conn read timeout")
				}
			} else {
				fmt.Println("no error")
				if n != 0 {
					json.Unmarshal(r.buffer[:n], &msg)
					fmt.Println("recieved", msg)
					r.C <- msg
				}
			}
		}
	}
}

func (r *Reciever) Start() {
	go r.startReciever()
}

func (r *Reciever) Stop() {
	r.kill <- struct{}{}
}

type senderPack struct {
	msg 	Message
	addr 	*net.UDPAddr
}

type Sender struct {
	C 		chan senderPack
	kill 	chan struct{}
	udpConn *net.UDPConn
	buffer 	[]byte
}

func NewSender(udpConn *net.UDPConn) *Sender {
	sndr := &Sender{make(chan senderPack, 100), make(chan struct{}), udpConn, make([]byte, 0, 1024)} 
	return sndr
}

func (s *Sender) startSender() {
	for {
		select {
		case <-s.kill:
			return
		case pack := <-s.C:
			buffer, _ := json.Marshal(pack.msg)
			fmt.Println("sending msg by sender", pack.msg)
			n, err := s.udpConn.WriteToUDP(buffer, pack.addr)
			fmt.Println(n, err)
		}
	}
}

func (s *Sender) Start() {
	go s.startSender()
}

func (s *Sender) Stop() {
	s.kill <- struct{}{}
}



type GossipNode struct {
	id 			int
	port 		int
	udpConn 	*net.UDPConn
	reciever 	*Reciever
	sender 		*Sender
	processor 	*nodeProcessor
}

func NewGossipNode (id int, port int, neighs []graph.Node) *GossipNode {
	return &GossipNode{
		id: 		id,
		port: 		port,
		udpConn: 	nil,
		reciever: 	nil,
		sender: 	nil,
		processor: 	newNodeProcessor(id, neighs),
	}
}

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
	gn.reciever = NewReciever(conn)
	gn.sender = NewSender(conn)
}

func (gn *GossipNode) Unbind() {
	gn.reciever.Stop()
	gn.sender.Stop()
	gn.udpConn.Close()
	logger.Println(time.Now().String(), "NODE", gn.id, "TRACE: port unbinded")
}

func (gn *GossipNode) putNewRumour(msg Message, netSize int) (exists bool) {
	return gn.processor.initNewMessage(msg, netSize)
}

func (gn *GossipNode) process(kill chan struct{}) {
	fmt.Println(gn)
	logger.Printf("[NODE %d] started processing", gn.id)
	gn.Bind()
	fmt.Println("binded")
	defer gn.Unbind()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	gn.reciever.Start()
	defer gn.reciever.Stop()
	gn.sender.Start()
	defer gn.sender.Stop()
	fmt.Println(gn)
	counter := 0
	fmt.Println("loop next")
	fmt.Println(len(gn.reciever.C))
	for {
		select {
		case <-kill:
			return
		case msg := <-gn.reciever.C:
			logger.Printf("[NODE %d] message recieved %s", gn.id, msg.String())
			gn.processor.processMsg(msg)
		case <-ticker.C:
			counter++
			msg, addr, empty := gn.processor.getRandomMsg()
			if !empty {
				logger.Printf("[NODE %d] sending to address %s message %s", gn.id, addr, msg)
				gn.sender.C <- senderPack{msg, addr}
			}
			msg, addr, empty = gn.processor.getRandomAck()
			if !empty {
				gn.sender.C <- senderPack{msg, addr}
			}
		}
	}
}


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

func (GN *GossipNet) Start() {
	//logger = initLogger("D:\\Documents\\sem7\\clouds_hpc\\task2_gossip\\")
	logger = initLogger("")
	logger.Println(time.Now().String(), " Start")
	for i:= 0; i < GN.size; i++ {
		go GN.nodes[i].process(GN.kill)
	}
	time.Sleep(time.Second)
}

func (GN *GossipNet) Stop() {
	for i := 0; i <= GN.size; i++ {
		time.Sleep(50 * time.Millisecond)
		GN.kill <- struct{}{}
	}
	closeLogger()
}

type errorString struct {
    s string
}

func (e *errorString) Error() string {
    return e.s
}

func (GN *GossipNet) MakeRumour(id int, msg Message) error {
	exists := GN.nodes[id].putNewRumour(msg, GN.size)
	if exists {
		return &errorString{"such message ID has been sent already"}
	}
	return nil
}
