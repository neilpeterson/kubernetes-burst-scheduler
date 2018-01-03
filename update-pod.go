package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	// Marshal JSON
	body, err := json.Marshal(pu)
	if err != nil {
		fmt.Println(err)
	}

	// HTTP Post to update node name.
	url := "http://localhost:8001/api/v1/namespaces/default/pods/" + podName + "/binding"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(response.Status))
}
