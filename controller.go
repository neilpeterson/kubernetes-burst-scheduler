// TODO - Pick back up on returning slice of nodes and burtsnode
// TODO - Cluster config and api updates.
// TODO - Refactor variable names, files.
// TODO - Readme.

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
	nodes           corev1.NodesGetter
}

func newNodeBurstController(client *kubernetes.Clientset, podInformer informercorev1.PodInformer) *nodeBurstController {

	c := &nodeBurstController{
		podGetter:       client.CoreV1(),
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "secretsync"),
		nodes:           client.CoreV1(),
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

	// Process work item
	err := c.processItem(key.(string))
	if err == nil {
		c.queue.Forget(key)
		return true
	}
	return true
}

func (c *nodeBurstController) processItem(key string) error {

	// Get pod, TODO - Update to use informer.GetIndexer().GetByKey(key)
	pod := c.getPod(strings.Split(key, "/")[1])

	if pod != nil {

		defaultScheduler := c.calculatePodPlacement(pod)

		if defaultScheduler {
			log.Println("Scheduling pod using default scheduler: " + pod.GetName())

			// Get node list
			n, _ := c.listNodes()

			// Get random node - TODO: update this to change to the default scheduler
			randomNode := getRandomNode(n)

			// Schedule on random node
			schedulePod(pod.GetName(), randomNode)

		} else {
			log.Println("Scheduling pod on burst node: " + pod.GetName())

			// Validate node - get node list
			n, bn := c.listNodes()

			// Validate burst node
			validNode := c.checkNode(n, bn)

			if validNode {
				schedulePod(pod.GetName(), burstNode)
			} else {
				log.Printf("%s%s%s", "Node: ", burstNode, " can not be found.")
			}
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
