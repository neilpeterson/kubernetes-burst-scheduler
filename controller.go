// TODO - figure out config file /  environment variable
// TODO - provide name for scheduler
// TODO - updated schedule on node to use default scheduler
// TODO - how will this function if I process one pod at a time?

package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var burstValue = 2

type nodeBurstController struct {
	podGetter       corev1.PodsGetter
	podLister       listercorev1.PodLister
	podListerSynced cache.InformerSynced
	queue           workqueue.RateLimitingInterface
}

// Node burst controller with an on add function
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

// Run node burst controller
func (c *nodeBurstController) Run(stop <-chan struct{}) {
	var wg sync.WaitGroup

	// Stop queue and workers
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

	// Here I am getting the pod using the split key.
	// TODO - update to use informer.GetIndexer().GetByKey(key)
	pod := c.getPod(strings.Split(key, "/")[1])

	if pod != nil {
		// log.Println("Start calculate pod placement")
		burst := c.calculatePodPlacement(pod)

		name := pod.GetName()

		if burst {
			// Burst schedule
			schedulePod(name, "virtual-kubelet-myaciconnector-linux")
		} else {
			// Use default scheduler
			schedulePod(name, "aks-nodepool1-23443254-0")
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

	// Get pod label, used to find all replicas for burst calculation
	podLabel := pod.GetLabels()["app"]

	rawPODS, _ := c.podLister.Pods("default").List(labels.Everything())

	//var scheduled int
	var notScheduled int

	for _, pod := range rawPODS {
		if pod.GetLabels()["app"] == podLabel {

			//if pod.Status.Phase == "Pending" {
			if pod.Spec.NodeName == "" {

				notScheduled++
				if notScheduled < burstValue {
					return true
				}

			} else {
				fmt.Println("Pod is allready schedled.")
				return false
			}

		} else {
			// Discard
			log.Println("Discard")
		}
		duration := time.Duration(1) * time.Second
		time.Sleep(duration)
	}

	// Return true to burst / false to use default
	return true
}
