package main

import (
	"fmt"
	"github.com/sokks/gossip"
	//"time"
)

func main() {
	gossipNet := gossip.InitNet(4)
	gossipNet.Start()
 	msg := gossip.Message{
		ID:		1,
		MsgType:"multicast",
		Sender:	0,
		Origin:	0,
		Data:	"initial message",
	} 
	gossipNet.MakeRumour(0, msg)
	fmt.Scanln()
	gossipNet.Stop()
	println("finished")
}