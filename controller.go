// TODO - figure out config file /  environment variable
// TODO - provide name for scheduler
// TODO - updated schedule on node to use default scheduler
// TODO - how will this function if I process one pod at a time?

package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		burst := calculatePodPlacement(pod)

		if burst {
			// Burst schedule
		} else {
			// Use default scheduler
		}
	}
	return nil
}

func (c *nodeBurstController) getPod(podName string) *v1.Pod {
	pod, _ := c.podGetter.Pods("default").Get(podName, metav1.GetOptions{})

	// TODO - filter out initial pods by namespace.
	// Return err
	// if err != nil {
	// 	log.Println(err)
	// }

	if (pod.Spec.SchedulerName == "test-scheduler") && (pod.Spec.NodeName == "") {
		log.Println("Pod " + podName + " needs to be scheduled.")
		return pod
	}
	return nil
}

func calculatePodPlacement(pod *v1.Pod) bool {
	labels := pod.GetLabels()["app"]
	log.Println(labels)

	// TODO:
	// Get all pods with app label
	// Calculate scheduled
	// Calculate scheduled on burst node
	// Return burstNode or default scheduler

	// Return true to burst / false to use default
	return true

}

// func (c *nodeBurstController) onAdd(obj interface{}) {

// 	// Get pods assigned to custom scheduler.
// 	pods, _ := c.getPods()

// 	// Get current state of pods (PendingSchedule vs. Scheduled).
// 	psch, sch := c.getCurrentState(pods)

// 	// Calcuate pod placement and schedule.
// 	calculatePodPlacement(psch, sch, pods)
// }

// Returns a slice of pods with custom scheduler and no assignment
//TODO - how will this function if I process one pod at a time?
// func (c *nodeBurstController) getPods() ([]*v1.Pod, error) {

// 	rawPODS, _ := c.podLister.Pods("Default").List(labels.Everything())
// 	pods := []*v1.Pod{}

// 	for _, pod := range rawPODS {
// 		if (pod.Spec.SchedulerName == "test-scheduler") && (pod.Spec.NodeName == "") {
// 			pods = append(pods, pod)
// 		}
// 	}
// 	log.Println(len(pods))
// 	return pods, nil
// }

// // Scheduler Calculation
// func (c *nodeBurstController) getCurrentState(pods []*v1.Pod) (int, int) {

// 	// Store app labels for calculation
// 	appLabel := map[string]bool{}

// 	PendingSchedule := 0
// 	Scheduled := 0

// 	// Add app label to map if not exsist
// 	for _, p := range pods {
// 		if appLabel[p.GetLabels()["app"]] {

// 		} else {
// 			appLabel[p.GetLabels()["app"]] = true
// 		}

// 		// Calculate allready scheduled, and need to schedule
// 		for _, pod := range pods {
// 			if appLabel[pod.GetLabels()["app"]] {
// 				if pod.Status.Phase == "Pending" {
// 					PendingSchedule++
// 				} else {
// 					Scheduled++
// 				}
// 			}
// 		}
// 	}
// 	return PendingSchedule, Scheduled
// }

// // Calculate placement and run function to schedule on node
// func calculatePodPlacement(psch int, sch int, pods []*v1.Pod) {

// 	newInt := 0

// 	if sch < burstValue {
// 		newInt = burstValue - sch

// 		for _, pod := range pods {
// 			log.Println(pod.GetName())
// 			if newInt > 0 {
// 				log.Println("Schedule on node..")
// 				schedulePod(pod.GetName(), "aks-nodepool1-42032720-0")
// 				newInt--
// 			} else {
// 				log.Println("Schedule on burst node..")
// 				schedulePod(pod.GetName(), "aks-nodepool1-42032720-2")
// 				newInt--
// 			}
// 		}
// 	}
// }
