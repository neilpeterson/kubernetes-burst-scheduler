package main

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *nodeBurstController) balancePods() {

	var node int
	var bnode int

	var podLabels map[string]string
	podLabels = make(map[string]string)

	rawPODS, _ := c.podLister.Pods("default").List(labels.Everything())

	for _, pod := range rawPODS {
		if pod.Spec.SchedulerName == *schedulerName {
			// TODO - not sure about this, element is just beeing overwritten
			// but each pod with identical label.
			// Lean more about maps..
			podLabels[pod.GetLabels()["app"]] = pod.GetName()
		}
	}

	for label := range podLabels {
		rawPODS, _ := c.podGetter.Pods("").List(metav1.ListOptions{LabelSelector: "app=" + label})

		for _, pod := range rawPODS.Items {
			if pod.Spec.NodeName == *burstNode {
				fmt.Println("on burst node")
				bnode++
			} else {
				fmt.Println("on node")
				node++
			}
		}
	}

	// HERE - NEED TO CALC Delete Algorythem and figure out how to select pod, etc..
	if node < *burstValue {
		c.podGetter.Pods("").Delete("opoppo", &metav1.DeleteOptions{})
	}
}
