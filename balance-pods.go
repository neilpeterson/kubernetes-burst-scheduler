package main

import "k8s.io/apimachinery/pkg/labels"

func (c *nodeBurstController) balancePods() {

	var labels2 []string

	rawPODS, _ := c.podLister.Pods("default").List(labels.Everything())

	for _, pod := range rawPODS {
		labels2 := append(labels2, pod.GetLabels()["app"])
	}
}
