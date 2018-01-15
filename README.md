# Kubernetes burst scheduler

Simple scheduler that will burst workload to a named node after a specified number of related pods have been started.

## Example use case:

You have a Kuebrentes cluster that consists of three nodes. Virtual Kublet has been configured to present Azure Container Instances as a virtual node on a Kubernetes cluster. The cluster looks like this:

```
NAME                                   STATUS    ROLES     AGE       VERSION
aks-nodepool1-34059843-0               Ready     agent     9h        v1.7.7
aks-nodepool1-34059843-1               Ready     agent     9h        v1.7.7
aks-nodepool1-34059843-2               Ready     agent     9h        v1.7.7
virtual-kubelet-myaciconnector-linux   Ready     agent     2m        v1.8.3
```

A batch processing routine automatically starts jobs on the Kuebrentes cluster to process some work. In general, 5  - 10 jobs are being processed / pods started at any given time. Occasionally an event occurs that temporarily increases this workload to 15 – 20 concurrent jobs. 

You would like to primarily run these jobs on the Kubernetes nodes in your cluster, however is the number of concurrently running pods increases above 10, these pods should be scheduled on Azure Container Instances by the virtual kublet node.

As the number of concurrently running jobs drops back under 10, pods will be scheduled on the Kuebrentes nodes.

## Deployment

Run the following to start the burst scheduler.

```
kubectl create -f https://raw.githubusercontent.com/neilpeterson/k8s-go-controller/master/manifest-files/burst-scheduler.yaml
```

Here is the manifest file that is run.

```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: gogurt
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: gogurt
    spec:
      containers:
      - name: kubectl-sidecar
        image: neilpeterson/kubectl-proxy-sidecar
      - name: gogurt
        image: neilpeterson/gogurt
```

## Execution

```
go run burst-scheduler --burstNode virtual-kubelet-myaciconnector-linux --burstValue 10
```

Arguments:

| Argument | Type | Description |
|---|---|---|
| burstNode | String | Node name of the burst node. This is the node on which pods are scheduled once the burstValue has been met. |
| burstValue | Int | Value that controls how many pods will be scheduled on Kubernetes nodes vs. burst node. |
| kubeConfig | Bool | Indicates that a kubernetes config file found at $KUBECONFIG is used for cluster discovery / auth. If not specified, it is assumed execution is occurring from a pod in the Kubernetes cluster. |

## TODO:

**Label Filter** - currenly using a loop to inventory and combine nodes with a common label. Update to filter the returned list. See this [doc for a sample](http://blog.kubernetes.io/2018/01/introducing-client-go-version-6.html), would this work here.

**Terminating pods** – filter these from scope. Not a big issue but can be problematic during demos / quick turn-a-rounds.

**Namespace** - currently 'default' is a non-configurable default. Update with a `--namespace` argument.

**Default Scheduler** - Update pod updater to use default scheduler when not in burst. Currently a random node from all nodes - the burst node is chosen for scheduling. I am not able to patch the pod scheduler property value.

**Pod node assignment** - Update this to use go client method, go client rest interface. Currently the node assignment is handled through direct api call (via kubectl proxy / sidecar). I was unable to use podInterface.update due to a non-updatable property (error below). Is there a way to complete through the go client, if not I see the that there is a REST interface, this is probably a better way to achieve this.

```
"aci-helloworld-4142002832-3l873" is invalid: spec: Forbidden: pod updates may not change fields other than `spec.containers[*].image`, `spec.initContainers[*].image`, `spec.activeDeadlineSeconds` or `spec.tolerations` (only additions to existing tolerations)
```

https://github.com/kubernetes/kubernetes/issues/24913