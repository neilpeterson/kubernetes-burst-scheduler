// TODO - figure out config file /  environment variable
// TODO - provide name for scheduler
// TODO - updated schedule on node to use default scheduler
// TODO - fix caching issue / mising go routine
// TODO - somethign is looping through and rescheduling all pods on node

package main

import (
	"log"
	"time"

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
				time.Sleep(2 * time.Second)
			},
		},
	)

	return c
}

// Run node burst controller
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

	// Get pods assigned to custom scheduler.
	pods, _ := c.getPods()

	// Get current state of pods (PendingSchedule vs. Scheduled).
	psch, sch := c.getCurrentState(pods)

	// Calcuate pod placement and schedule.
	calculatePodPlacement(psch, sch, pods)
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
	log.Println(len(pods))
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

		// Calculate allready scheduled, and need to schedule
		for _, pod := range pods {
			if appLabel[pod.GetLabels()["app"]] {
				if pod.Status.Phase == "Pending" {
					PendingSchedule++
				} else {
					Scheduled++
				}
			}
		}
	}
	return PendingSchedule, Scheduled
}

// Calculate placement and run function to schedule on node
func calculatePodPlacement(psch int, sch int, pods []*v1.Pod) {

	newInt := 0

	if sch < burstValue {
		newInt = burstValue - sch

		for _, pod := range pods {
			log.Println(pod.GetName())
			if newInt > 0 {
				log.Println("Schedule on node..")
				schedulePod(pod.GetName(), "aks-nodepool1-42032720-0")
				newInt--
			} else {
				log.Println("Schedule on burst node..")
				schedulePod(pod.GetName(), "aks-nodepool1-42032720-2")
				newInt--
			}
		}
	}
}
