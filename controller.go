package main

import (
	"log"

	informercorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type nodeBurstController struct {
	podGetter       corev1.PodsGetter
	podLister       listercorev1.PodLister
	podListerSynced cache.InformerSynced
}

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

// Scraped from JOES VIDEO BEFORE UPDATE
func (c *nodeBurstController) Run(stop <-chan struct{}) {
	log.Print("Waiting for cache sync")
	if !cache.WaitForCacheSync(stop, c.podListerSynced) {
		log.Print("Timed out while waiting for cache")
		return
	}

	log.Print("caches are synced")

	log.Print("Waiting for stop signal")
	<-stop
	log.Print("Recieved stop singnal")
}

func (c *nodeBurstController) onAdd(obj interface{}) {
	key, _ := cache.MetaNamespaceKeyFunc(obj)
	log.Printf("onAdd: %v", key)
}
