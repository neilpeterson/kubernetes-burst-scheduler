
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
kubectl create -f <add file>
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
|---|---|
| burstNode | String | Node name of the burst node. This is the node on which pods are scheduled once the burstValue has been met. |
| burstValue | Int | Value that controls how many pods will be scheduled on Kubernetes nodes vs. burst node. |
| kubeConfig | Bool | Indicates that a kubernetes config file found at $KUBECONFIG is used for cluster discovery / auth. If not specified, it is assumed execution is occurring from a pod in the Kubernetes cluster. |

## TODO:

Terminating pods – filter these from scope. Not a big issue but can be problematic during demos / quick turn-a-rounds.

Namespace - currently 'default' is a non-configurable default. Update with a `--namespace` argument.

Default Scheduler - Update pod updater to use default scheduler when not in burst. Currently a random node from all nodes - the burst node is chosen for scheduling.

API Authentication - go client is working well, however unsure how to handle direct api call. Currently using side car / kubectl proxy (recommended in docs). Can I a. auth / raw rest call through the go client. b. See next TODO.

Update node - currently the node assignment is handled through direct api call. I was unable to used pod.update due to a non-updatable property (target.name). Is there a way to complete through the go client? This would produce neater code and remove the need for the side car container.





