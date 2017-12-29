// TODO1 - Complete scheduling
// TODO2 - Figure out cache sync / missing go routine
// TODO3 - Figure out incorect pod integer / incorect ammount

package main

import (
	"log"

	informercorev1 "k8s.io/client-go/informers/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var burstValue = 2

type nodeBurstController struct {
	podGetter       corev1.PodsGetter
	podLister       listercorev1.PodLister
	podListerSynced cache.InformerSynced
}

// Node burst controller with an on add function
func newNodeBurstController(client *kubernetes.Clientset, podInformer informercorev1.PodInformer) *nodeBurstController {

	c := &nodeBurstController{
		podGetter:       client.CoreV1(),
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
	}

	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.onAdd(obj)
			},
		},
	)

	return c
}

// Run node burst controller
// TODO2 - Figure out cache sync / missing go routine
func (c *nodeBurstController) Run(stop <-chan struct{}) {

	log.Print("waiting for cache sync")

	if !cache.WaitForCacheSync(stop, c.podListerSynced) {
		log.Print("Timed out while waiting for cache")
		return
	}
	log.Println("Caches are synced")

	log.Print("Waiting for stop singnal")
	<-stop
	log.Print("Recieved stop singnal")
}

func (c *nodeBurstController) onAdd(obj interface{}) {

	// Get pods using custom scheduler.
	pods, _ := c.getPods()

	// Get current state of pods (PendingSchedule vs. Scheduled).
	psch, sch := c.getCurrentState(pods)

	// Calcuate pod placement.
	calculatePodPlacement(psch, sch, pods)

	// Schedule pods on nodes.
	// schedulePod("podName", "nodeName")
}

// Returns a slice of pods with custom scheduler and no assignment
func (c *nodeBurstController) getPods() ([]*v1.Pod, error) {

	rawPODS, _ := c.podLister.Pods("").List(labels.Everything())
	pods := []*v1.Pod{}

	for _, pod := range rawPODS {
		if (pod.Spec.SchedulerName == "test-scheduler") && (pod.Spec.NodeName == "") {
			pods = append(pods, pod)
		}
	}

	return pods, nil
}

// Scheduler Calculation
func (c *nodeBurstController) getCurrentState(pods []*v1.Pod) (int, int) {

	// Store app labels for calculation
	appLabel := map[string]bool{}

	PendingSchedule := 0
	Scheduled := 0

	// Add app label to map if not exsist
	for _, p := range pods {
		if appLabel[p.GetLabels()["app"]] {

		} else {
			appLabel[p.GetLabels()["app"]] = true
		}

		// Filter pods on app label TODO - how to use label selector
		// TODO3 - Figure out incorect pod integer / incorect ammount
		filterPODS, _ := c.podLister.Pods("").List(labels.Everything())

		// Calculate allready scheduled, and need to schedule
		for _, pod := range filterPODS {
			if appLabel[pod.GetLabels()["app"]] {
				if pod.Status.Phase == "Pending" {
					// Incriment to Schedule
					log.Println("To Schedule")
					PendingSchedule++
				} else {
					// Incriment scheduled
					log.Println("Allready Schedule")
					Scheduled++
				}
			}
		}
	}
	return PendingSchedule, Scheduled
}

func calculatePodPlacement(psch int, sch int, pod []*v1.Pod) {

	// TODO1 - Complete scheduling
	// Write equation

	log.Println(psch)
	log.Println(sch)

	// Psudo Code
	// Need to bring in a list of pods to schedule []pod.
	// Perform calculatios
	// Remove node from slice until slice is empty

	// if psch(3) + sch(0) > burstvalue(2) {
	// 	intNode(2) = burstValue(2) - sch(0)
	// 	intBurst(1) = psch(3) - burstValue(2)
	// 	Do while intNode(2) >= 0 {
	// 		schedule node
	// 		remove pod from slice - https://stackoverflow.com/questions/25025409/delete-element-in-a-slice
	// 		-- intNode(1),(0)
	// 	}

	// 	do while intBurst(1) >= 0 {
	// 		scheudle ACI
	// 		remove pod from slice - https://stackoverflow.com/questions/25025409/delete-element-in-a-slice
	// 		--intBurst(0)
	// 	}
	// }

}
