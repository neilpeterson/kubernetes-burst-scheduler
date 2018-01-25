package main

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *nodeBurstController) balancePods() {

	var podLabels map[string]string
	podLabels = make(map[string]string)

	rawPODS, _ := c.podLister.Pods("default").List(labels.Everything())

	for _, pod := range rawPODS {
		if pod.Spec.SchedulerName == *schedulerName {
			podLabels[pod.GetLabels()["app"]] = pod.GetName()
		}
	}

	for label := range podLabels {
		n, bn := c.getNodeWeight(label)
		c.reBalancePods(n, bn)
	}
}

func (c *nodeBurstController) reBalancePods(n []v1.Pod, bn []v1.Pod) {

	// Stop node, which when backed by replicas set, will reschedule appropriatley.
	if len(n) < *burstValue {
		i := len(bn)

		for i != 0 {
			err := c.podGetter.Pods("default").Delete(bn[0].GetName(), &metav1.DeleteOptions{})
			if err != nil {
				fmt.Println(err)
			}
			i--
		}
	}
}
