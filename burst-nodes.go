package main

import (
	"log"
	"math/rand"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *nodeBurstController) listNodes() ([]string, bool) {

	var nodeList []string
	var valid bool

	// Build list of nodes.
	nodes, _ := c.nodes.Nodes().List(metav1.ListOptions{})

	// Validate burst node and remove from list.
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

// Get random node from list.
func getRandomNode(nodeList []string) string {

	n := rand.Int() % len(nodeList)
	node := nodeList[n]

	log.Println("Random Node: " + node)
	return string(node)
}

// For a given label, retrun count of pods on node and burst node.
func (c *nodeBurstController) getNodeWeight(podLabel string) ([]v1.Pod, []v1.Pod) {

	var node []v1.Pod
	var bnode []v1.Pod

	rawPODS, _ := c.podGetter.Pods("").List(metav1.ListOptions{LabelSelector: "app=" + podLabel})

	for _, pod := range rawPODS.Items {
		if (pod.DeletionTimestamp == nil) && (pod.Spec.NodeName != "") {
			if pod.Spec.NodeName == *burstNode {
				bnode = append(bnode, pod)
			} else {
				node = append(node, pod)
			}
		}
	}
	return node, bnode
}
