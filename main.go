package main

import (
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	// Initilization information - package rest
	var (
		config *rest.Config
		err    error
	)

	kubeconfig := os.Getenv("KUBECONFIG")

	// Authentication / connection object - package tools/clientcmd
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v", err)
		os.Exit(1)
	}

	// Kubernetes client - package kubernetes
	clientset := kubernetes.NewForConfigOrDie(config)

	// Shared Informer - package informers
	sharedInformers := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	nodeBurstController := newNodeBurstController(clientset, sharedInformers.Core().V1().Pods())
	sharedInformers.Start(nil)
	nodeBurstController.Run(nil)
}
