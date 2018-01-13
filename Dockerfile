## Quick job to test insdie pod
## TODO - build proper Dockerfile

FROM golang:1.8

RUN go get k8s.io/client-go/...

ADD . /app

WORKDIR /app

ENTRYPOINT ["/usr/local/go/bin/go", "run", "main.go", "get-nodes.go", "update-pod.go", "controller.go", "--burstNode", "aks-nodepool1-34059843-2" ]