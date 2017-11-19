package gossip

import (
	"fmt"
	"net"
	"gitlab.com/n-canter/graph"
	"strconv"
	"strings"
)

type nodeProcessor struct {
	myID 		int						// unique id of processor in the Net
	neighbours 	map[int]*net.UDPAddr 	// map[nodeID]nodeAddr
	msgIDs 		[]int					// slice of already received message IDs
	ackIDs 		map[int][]int			// map[msgID]slice of node IDs sent ack with msgID
	msgQueue 	*messageQueue			// queue of messages to send
	ackQueue 	*messageQueue			// queue of acks to send
	acks 		map[int][]bool			// map[msgID](slice[nodeID]=true/false)
										//	 note: if msgID in keys of this map then it was 
										// 		   send by this node and is tracked by it
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

// initNewMessage puts msg in the message queue and allocates resources for tracking it
//
// TODO: generate unique message id internally
// TODO: use mutex to avoid data races (though such races have impact 
// 		 only for delay of including message in the processing)
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
	if !existingID(msgId) { // If such msgId is already known to this node then error is generated
							// but it doesn't garantee that there are no sucj msgId in the whole Net.
							// If it's already exists it will be ignored by nodes or can be processed 
							// incorrectly.
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
		if !hasKey {
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
		fmt.Println(p.myID, msg.ID, p.ackIDs[msg.ID])
		if !alreadyReceivedAck(msg.ID, msg.Origin) {
			memorizeAckID(msg.ID, msg.Origin)
			if initedByMe(msg.ID) {
				fmt.Println(p.myID, msg.ID, "inited by me")
				writeAck(msg.ID, msg.Origin)
				logger.Printf("[NODE %d] [MESSAGE %d ACKED BY NODE %d] acks for message: %s \n", p.myID, msg.ID, msg.Origin, boolSliceToString(p.acks[msg.ID]))
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
