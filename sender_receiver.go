package gossip

import (
	"net"
	"encoding/json"
	"time"
)

// Receiver is a non-blocking reader from UDP connection.
// It has a seporate listening goroutine which puts received
// data in the channel. 
type Receiver struct {
	C chan Message
	kill chan struct{}
	udpConn *net.UDPConn
	buffer []byte
}

// NewReceiver constracts a new Receiver object assosiated with udpConn.
func NewReceiver(udpConn *net.UDPConn) *Receiver {
	rcvr := &Receiver{make(chan Message, 100), make(chan struct{}), udpConn, make([]byte, 1024)} 
	return rcvr
}

func (r *Receiver) startReceiver() {
	timeout := time.Duration(100 * time.Millisecond)
	for {
		select {
		case <-r.kill:
			return
		default:
			msg := Message{}
			r.udpConn.SetReadDeadline(time.Now().Add(timeout))
			n, _, err := r.udpConn.ReadFromUDP(r.buffer)
			if err != nil {
				if e, ok := err.(net.Error); !ok || !e.Timeout() {
					panic(err)
				} else { continue }
			} else {
				if n != 0 {
					json.Unmarshal(r.buffer[:n], &msg)
					r.C <- msg
				}
			}
		}
	}
}

// Start launches the receiver
func (r *Receiver) Start() {
	go r.startReceiver()
}

// Stop sends stop signal to receiver
func (r *Receiver) Stop() {
	r.kill <- struct{}{}
}

type senderPack struct {
	msg 	Message
	addr 	*net.UDPAddr
}

// Sender is a writer to UDP connection.It has a seporate 
// goroutine which gets task from a channel and sends it. 
// Task is a pair of message and its recipient address.
type Sender struct {
	C 		chan senderPack
	kill 	chan struct{}
	udpConn *net.UDPConn
	buffer 	[]byte
}

// NewSender constructs new sender object assosiated with udpConn.
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
			_, err := s.udpConn.WriteToUDP(buffer, pack.addr)
			if err != nil {
				_ = err
			}
		}
	}
}

// Start launches the sender
func (s *Sender) Start() {
	go s.startSender()
}

// Stop sends stop signal to the sender
func (s *Sender) Stop() {
	s.kill <- struct{}{}
}
