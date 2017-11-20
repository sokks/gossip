package main

import (
	"strconv"
	"bytes"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
	"time"
    
    "github.com/sokks/gossip"
    "gitlab.com/n-canter/graph" 
)

const (
    nExperiments = 10
)
func main() {
    fmt.Println("args:", os.Args)
    if len(os.Args) < 3 {
        fmt.Println("Not enough argumets: directory paths needed.")
        return
    }
    logDir := os.Args[1]
    dataDir := os.Args[2]
    dataFile := "test_loss.data"
    g := graph.Generate(50, 5, 7, 9080)
    probabilities := []string{"0.1", "0.2", "0.3", "0.4", "0.5"}
    datafile, err := os.OpenFile(filepath.Join(dataDir, dataFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
    if err != nil {
        fmt.Println("Can't open data file")
        return
    }
    defer datafile.Close()
    fmt.Println("Start collecting data, wait please")
    datafile.WriteString("10\n")
    datafile.WriteString("0.0\n")
    fmt.Print("Probability 0.0 ...")
    for i:= 0; i < nExperiments; i++ {
        doOneTest(g, logDir, datafile)
        fmt.Print(".....")
    }
    datafile.WriteString("\n")
    fmt.Println("")
    for _, p := range probabilities {
        if !ruleProbability(p, false) {
            break
        }
        datafile.WriteString(p + "\n")
        fmt.Print("Probability ", p, " ...")
        for i := 0; i < nExperiments; i++ {
            doOneTest(g, logDir, datafile)
            fmt.Print(".....")
        }
        datafile.WriteString("\n")
        fmt.Println("")
        ruleProbability(p, true)
    }
}

func ruleProbability(p string, toDel bool) bool {
    cmdArgs :=  []string{"-A", "INPUT", "-p", "udp", "-i", "lo", "-m", "statistic", 
							"--mode", "random", "--probability", "0.0", "-j", "DROP"} 
							// 0, 11 is variable
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
    if er.Len() > 0 {
        fmt.Println(er.String())
        return false
    }
    return true
}

func doOneTest(g graph.Graph, logDir string, datafile *os.File) {
    gossipNet := gossip.InitNetFromGraph(g, 100 * time.Millisecond)
    feedback := gossipNet.SetTestMode()
    gossipNet.Start(logDir)
    msg := gossip.Message{
        ID:        1,
        MsgType:   "multicast",
        Sender:    0,
        Origin:    0,
        Data:      "initial message",
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
    datafile.WriteString(strconv.Itoa(n) + " ")
}
