package main

import (
	"log"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// TODO - add function to validate node exsists
var burstNode = "aks-nodepool1-42032720-2"
var burstValue = 2

type nodeBurstController struct {
	podGetter       corev1.PodsGetter
	podLister       listercorev1.PodLister
	podListerSynced cache.InformerSynced
	queue           workqueue.RateLimitingInterface
}

func newNodeBurstController(client *kubernetes.Clientset, podInformer informercorev1.PodInformer) *nodeBurstController {

	c := &nodeBurstController{
		podGetter:       client.CoreV1(),
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "secretsync"),
	}

	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					c.queue.Add(key)
				}
			},
		},
	)

	return c
}

func (c *nodeBurstController) Run(stop <-chan struct{}) {
	var wg sync.WaitGroup

	defer func() {
		log.Println("Shutting down queue.")
		c.queue.ShutDown()

		log.Println("Shutting down worker")
		wg.Wait()

		log.Println("Workers are all done.")
	}()

	log.Print("waiting for cache sync")
	if !cache.WaitForCacheSync(stop, c.podListerSynced) {
		log.Print("Timed out while waiting for cache")
		return
	}
	log.Println("Caches are synced")

	go func() {
		wait.Until(c.runWorker, time.Second, stop)
		wg.Done()
	}()

	log.Print("Waiting for stop singnal")
	<-stop
	log.Print("Recieved stop singnal")
}

func (c *nodeBurstController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *nodeBurstController) processNextWorkItem() bool {

	// Pull work item from queue.
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	// Do the work
	err := c.processItem(key.(string))
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	return true
}

func (c *nodeBurstController) processItem(key string) error {

	// TODO - Update to use informer.GetIndexer().GetByKey(key)
	pod := c.getPod(strings.Split(key, "/")[1])

	if pod != nil {

		defaultScheduler := c.calculatePodPlacement(pod)

		if defaultScheduler {
			log.Println("Scheduling pod using default scheduler: " + pod.GetName())
			schedulePod(pod.GetName(), "aks-nodepool1-42032720-0")
		} else {
			log.Println("Scheduling pod on burst node: " + pod.GetName())
			schedulePod(pod.GetName(), burstNode)
		}
	}
	return nil
}

func (c *nodeBurstController) getPod(podName string) *v1.Pod {
	pod, _ := c.podGetter.Pods("default").Get(podName, metav1.GetOptions{})

	if (pod.Spec.SchedulerName == "test-scheduler") && (pod.Spec.NodeName == "") {
		return pod
	}
	return nil
}

func (c *nodeBurstController) calculatePodPlacement(pod *v1.Pod) bool {

	var scheduled int
	var track int

	// Get all pods with matching label
	podLabel := pod.GetLabels()["app"]
	rawPODS, _ := c.podLister.Pods("default").List(labels.Everything())

	// Calculate placement
	for _, pod := range rawPODS {
		if pod.GetLabels()["app"] == podLabel {
			if pod.Spec.NodeName != "" {
				scheduled++
			}
		}
	}
	track = burstValue - scheduled
	if track > 0 {
		// Default scheduler
		return true
	}
	// Burst node
	return false
}
