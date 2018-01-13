// TODO - filter terminating pods.
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

// Start informer, build cache
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

// Start worker - watched queue, call processor
func (c *nodeBurstController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// Pulls item from queue
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

// Process item - work starts here
func (c *nodeBurstController) processItem(key string) error {

	// Get pod, TODO - Update to use informer.GetIndexer().GetByKey(key)
	pod := c.getPod(strings.Split(key, "/")[1])

	if pod != nil {

		// Determine if the burst node is needed
		defaultScheduler := c.calculatePodPlacement(pod)

		if defaultScheduler {
			log.Println("Scheduling pod using default scheduler: " + pod.GetName())

			// Get node list
			n, _ := c.listNodes()

			// Get random node
			randomNode := getRandomNode(n)

			// Schedule pod on random node
			schedulePod(pod.GetName(), randomNode)

		} else {
			log.Println("Scheduling pod on burst node: " + pod.GetName())

			// Validate burst nodes exsists
			_, bn := c.listNodes()

			if bn {
				// Schedule pod on random node
				schedulePod(pod.GetName(), *burstNode)
			} else {
				log.Printf("%s%s%s", "Node: ", *burstNode, " can not be found.")
			}
		}
	}
	return nil
}

// Get a single pod by name
func (c *nodeBurstController) getPod(podName string) *v1.Pod {
	pod, _ := c.podGetter.Pods("default").Get(podName, metav1.GetOptions{})

	if (pod.Spec.SchedulerName == "test-scheduler") && (pod.Spec.NodeName == "") {
		return pod
	}
	return nil
}

// Calculates state for pods with matching labels
// Determins if the burst node is needed
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
	track = *burstValue - scheduled
	if track > 0 {
		// Default scheduler
		return true
	}
	// Burst node
	return false
}
