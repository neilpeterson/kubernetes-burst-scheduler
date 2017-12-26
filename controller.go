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

type nodeBurstController struct {
	podInterface    corev1.PodInterface
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
// TODO - What is going on here, I have a channel but no go routine.
// Is is here that is causing the unneciary initilization
// Disect this function more closley
func (c *nodeBurstController) Run(stop <-chan struct{}) {

	log.Print("waiting for cache sync")

	if !cache.WaitForCacheSync(stop, c.podListerSynced) {
		log.Print("Timed out while waiting for cache")
		return
	}

	<-stop
	log.Print("Recieved stop singnal")
}

func (c *nodeBurstController) onAdd(obj interface{}) {
	// key, _ := cache.MetaNamespaceKeyFunc(obj)
	// log.Printf("onAdd: %v", key)
	pods, _ := c.getPods()

	// FOR TEST - Added a reciever function and calling it here.
	// Not sure if I need to pass the c struct
	c.calculateBurst(pods)
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

// Scheduler Calculation - TODO - do I need the recieve here?
func (c *nodeBurstController) calculateBurst(pods []*v1.Pod) {

	// Store app labels for calculation
	appLabel := map[string]bool{}

	toSchedule := 0
	allreadyScheduled := 0

	// Add app label to map if not exsist
	for _, p := range pods {
		if appLabel[p.GetLabels()["app"]] {

		} else {
			appLabel[p.GetLabels()["app"]] = true
		}

		// Filter pods on app label TODO - how to use label selector
		filterPODS, _ := c.podLister.Pods("").List(labels.Everything())

		// Calculate allready scheduled, and need to schedule
		for _, pod := range filterPODS {
			if appLabel[pod.GetLabels()["app"]] {
				if pod.Status.Phase == "Pending" {
					// Incriment to Schedule
					log.Println("To Schedule")
					toSchedule++
				} else {
					// Incriment scheduled
					log.Println("Allready Schedule")
					allreadyScheduled++
				}
			}
		}
	}
}

// Schedule pods
func schedulePods() {
	// Scheudle PODS
}
