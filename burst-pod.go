package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func schedulePod(podName string, nodeName string) {

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
		fmt.Println(err)
	}

	// Assign pod to node.
	// TODO - update to used go client method or rest client.
	url := "http://localhost:8001/api/v1/namespaces/default/pods/" + podName + "/binding"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	log.Println(resp.Status)
}
