package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type target struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type metadata struct {
	Name string `json:"name"`
}

type podUpdate struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   metadata `json:"metadata"`
	Target     target   `json:"target"`
}

func (c *nodeBurstController) schedulePod(podName string, nodeName string) {

	pu := &podUpdate{
		APIVersion: "v1",
		Kind:       "Binding",
		Metadata: metadata{
			Name: podName,
		},
		Target: target{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       nodeName,
		},
	}

	body, err := json.Marshal(pu)
	if err != nil {
		log.Println(err)
	}

	uri := "http://kubernetes/api/v1/namespaces/default/pods/" + nodeName + "/binding"

	err = c.rest.Post().RequestURI(uri).Body(body).Do().Error()
	if err != nil {
		fmt.Println(err)
	}
}
