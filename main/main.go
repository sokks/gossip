package main

import (
	"time"
	"fmt"
	"github.com/sokks/gossip"
)

func main() {
	logDir := "D:/Documents/sem7/cloud_hpc"
	gossipNet := gossip.InitNet(10, time.Second)
	gossipNet.Start(logDir)
	fmt.Println("Gossip simulation startted, see the log")
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
	fmt.Println("Simulation finished")
}