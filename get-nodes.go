package main

import (
	"fmt"
	"math/rand"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *nodeBurstController) listNodes() ([]string, string) {

	var nodeList []string
	var bn string

	nodes, _ := c.nodes.Nodes().List(metav1.ListOptions{})

	for _, n := range nodes.Items {
		if n.GetName() != burstNode {
			nodeList = append(nodeList, n.GetName())
		} else {
			bn = string(n.GetName())
		}
	}
	fmt.Println("BN: " + bn)
	fmt.Println(len(nodeList))
	return nodeList, bn
}

func (c *nodeBurstController) checkNode(nodeList []string, nodeName string) bool {
	fmt.Println("Checking for node: " + nodeName)
	for _, n := range nodeList {
		fmt.Println("N: " + n)
		if n == nodeName {
			return true
		}
	}
	return false
}

func getRandomNode(nodeList []string) string {

	// Get random node
	n := rand.Int() % len(nodeList)
	node := nodeList[n]

	fmt.Println("Random Node: " + node)
	return string(node)
}
