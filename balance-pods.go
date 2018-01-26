package main

import (
	"log"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *nodeBurstController) balancePods() {

	var pl map[string]string
	pl = make(map[string]string)

	rawPODS, _ := c.podLister.Pods(*namespace).List(labels.Everything())

	// Get inventory label for applicable pods.
	for _, pod := range rawPODS {
		if pod.Spec.SchedulerName == *schedulerName {
			pl[pod.GetLabels()["app"]] = pod.GetName()
		}
	}

	// For each label, balance pods.
	for label := range pl {
		n, bn := c.getNodeWeight(label)
		c.stopPods(n, bn)
	}
}

func (c *nodeBurstController) stopPods(n []v1.Pod, bn []v1.Pod) {

	// Stop node, which when backed by replicas set, will reschedule appropriately.
	if len(n) < *burstValue {
		i := len(bn)

		for i != 0 {
			err := c.podGetter.Pods(*namespace).Delete(bn[0].GetName(), &metav1.DeleteOptions{})
			if err != nil {
				log.Println(err)
			}
			i--
		}
	}
}
