package main

import (
	"time"
	"fmt"
	"github.com/sokks/gossip"
	"gitlab.com/n-canter/graph"
	"os/exec"
	"bytes"
)

func ruleProbability(p string, toDel bool) bool {
	cmdArgs :=  []string{"-A", "INPUT", "-p", "udp", "-i", "lo", "-m", "statistic", 
							"--mode", "random", "--probability", "0.0", "-j", "DROP"} // 0, 11 is variable
	var out bytes.Buffer
	var er bytes.Buffer
	if toDel {
		cmdArgs[0] = "-D"
	}
	cmdArgs[11] = p
	cmd := exec.Command("iptables", cmdArgs...)
	out.Reset()
	er.Reset()
	cmd.Stdout = &out
	cmd.Stderr = &er
	cmd.Run()
	//fmt.Println(out)
	if er.Len() > 0 {
		fmt.Println(er.String())
		return false
	}
	return true
}

func main() {
	//logDir := "D:/Documents/sem7/cloud_hpc"
	logDir := ""
	g := graph.Generate(50, 5, 7, 9080)
	probabilities := []string{"0.1", "0.2", "0.3", "0.4", "0.5"}
	fmt.Println("Probability: 0.0")
	doOneTest(g, logDir)
	for _, p := range probabilities {
		if !ruleProbability(p, false) {
			break
		}
		fmt.Println("Probability:", p)
		doOneTest(g, logDir)
		ruleProbability(p, true)
	}
}

func doOneTest(g graph.Graph, logDir string) {
	gossipNet := gossip.InitNetFromGraph(g, 100 * time.Millisecond)
	feedback := gossipNet.SetTestMode()
	gossipNet.Start(logDir)
	msg := gossip.Message{
		ID:		1,
		MsgType:"multicast",
		Sender:	0,
		Origin:	0,
		Data:	"initial message",
	} 
	gossipNet.MakeRumour(0, msg)
	var n int
	Loop:
	for {
		select {
		case n = <-feedback:
			break Loop
		}
	}
	gossipNet.Stop()
	fmt.Println("Result:", n)
}

func main1() {
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