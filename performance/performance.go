package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	fmt.Println("performance is great")
	fmt.Println("Args:", os.Args)
	if len(os.Args) < 6 {
		fmt.Println("Not enough arguments")
		return
	}
	NofNodes, _ := strconv.Atoi(os.Args[1])
    BasePort, _ := strconv.Atoi(os.Args[2])
    MinDeg, _   := strconv.Atoi(os.Args[3])
    MaxDeg, _   := strconv.Atoi(os.Args[4])
	TTL, _      := strconv.Atoi(os.Args[5])
	fmt.Println(NofNodes, BasePort, MinDeg, MaxDeg, TTL)
}