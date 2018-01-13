package main

import (
	"fmt"
	"math/rand"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *nodeBurstController) listNodes() ([]string, bool) {

	var nodeList []string
	var valid bool

	// Get all nodes
	nodes, _ := c.nodes.Nodes().List(metav1.ListOptions{})

	// Remove burst node from list
	// Validate burst node exsists
	for _, n := range nodes.Items {
		if n.GetName() != *burstNode {
			nodeList = append(nodeList, n.GetName())
		}
		if n.GetName() == *burstNode {
			valid = true
		}
	}

	if valid {
		return nodeList, true
	}
	return nodeList, false
}

// Get random node from list - burst node
// TODO - update to assign the default scheduler to pod
func getRandomNode(nodeList []string) string {

	n := rand.Int() % len(nodeList)
	node := nodeList[n]

	fmt.Println("Random Node: " + node)
	return string(node)
}
