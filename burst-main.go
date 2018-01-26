package main

import (
	"flag"
	"log"
	"os"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var namespace = flag.String("namespace", "default", "Kubernetes namespace.")
var burstNode = flag.String("burstNode", "", "Name of node onto which pods burst schedule.")
var burstValue = flag.Int("burstValue", 2, "Count of pods after which the burst node is scheduled.")
var kubeConfig = flag.Bool("kubeConfig", false, "Use a config file found at $KUBECONFIG for cluster authentication")
var schedulerName = flag.String("schedulerName", "burst-scheduler", "The name of the scheduler, this will match the named scheduler when deploying pods. The default value os burst-scheduler.")

func main() {

	var config *rest.Config
	var err error

	flag.Parse()

	if *kubeConfig {
		log.Println("Authenticate using config file found at $KUBECONFIG")
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig != "" {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		} else {
			log.Println("$KUBECONFIG path not set.")
			os.Exit(154)
		}
	} else {
		log.Println("In cluster authentication using go client.")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Println(err)
	}

	// Kubernetes client
	clientset := kubernetes.NewForConfigOrDie(config)

	// Shared Informer
	sharedInformers := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	nodeBurstController := newNodeBurstController(clientset, sharedInformers.Core().V1().Pods())
	sharedInformers.Start(nil)
	nodeBurstController.Run(nil)
}
