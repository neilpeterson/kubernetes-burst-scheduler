// https://github.com/squat/jupyter-operator/blob/9c7306f5cc37709275b7147b4fbb2863e829ef9a/pkg/k8sutil/pod.go

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
func (c *nodeBurstController) Run(stop <-chan struct{}) {

	log.Print("Waiting for cache sync")

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

	calculateBurst(pods)
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

// Calculate pod / node placement
func calculateBurst(pods []*v1.Pod) {

	// Used to store app labels for calculation
	// Using a map here instead of slice of strings for ease of content validation
	appLabel := map[string]bool{}

	// Check map for exsistance of app label
	// If it has not been added, add it
	for _, p := range pods {
		if appLabel[p.GetLabels()["app"]] {

		} else {
			appLabel[p.GetLabels()["app"]] = true
		}

		// Print out the map / debugging
		for m := range appLabel {
			log.Println(m)
		}
	}
}

// Schedule pods
func schedulePods() {

}
