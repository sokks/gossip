package gossip

import (
	"gitlab.com/n-canter/graph"
	"fmt"
)

func DoSomething() {
	g := graph.Generate(10, 1, 3, 9080)
	for i := 0; i < 10; i++ {
		node, _ := g.GetNode(i) // вернул ноду №10 тоже, но без соседей
		neighbours, _ := g.Neighbors(i)
		fmt.Println(node.String(), node.Port(), neighbours)
	}
}