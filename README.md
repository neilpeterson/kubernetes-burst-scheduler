# Kubernetes burst scheduler

Simple scheduler that will burst workload to a named node after a specified number of related pods have been started.

## Attribution

This is my first go project and first exposure to the Kubernetes go client. Throughout these two resources have been invaluable. Many thanks to the contributing teams.

- [Joe Beda controller sample](https://github.com/jbeda/tgik-controller)
- [Tu Nguyen kubewatch example / blog](https://engineering.bitnami.com/articles/kubewatch-an-example-of-kubernetes-custom-controller.html)

## Example use case:

You have a Kuebrentes cluster that consists of three nodes. Virtual Kublet has been configured to present Azure Container Instances as a virtual node on a Kubernetes cluster. The cluster looks like this:

```
NAME                                   STATUS    ROLES     AGE       VERSION
aks-nodepool1-34059843-0               Ready     agent     9h        v1.7.7
aks-nodepool1-34059843-1               Ready     agent     9h        v1.7.7
aks-nodepool1-34059843-2               Ready     agent     9h        v1.7.7
virtual-kubelet-myaciconnector-linux   Ready     agent     2m        v1.8.3
```

A batch processing routine automatically starts jobs on the Kuebrentes cluster. In general, 5 - 10 of these jobs/pods are running at any given time. Occasionally an event occurs that temporarily increases this workload to 15 – 20 concurrent jobs/pods. 

You would like to primarily run these jobs on the Kubernetes nodes, however when the running jobs/pods increases above 10, these pods should be scheduled on Azure Container Instances by the virtual kublet node.

## Deployment

The following manifest can be used to start the scheduler. Update `<node-name>` with the name of the burst node, and `<integer>` with the burst value. See the next section for details on all possible arguments.

```yaml
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: burst-scheduler
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: burst-scheduler
    spec:
      containers:
      - name: kubectl-sidecar
        image: neilpeterson/kubectl-proxy-sidecar
      - name: burst-scheduler
        image: neilpeterson/burst-scheduler:v1
        args: ["--burstNode", "<node-name>", "--burstValue", "<integer>"]
```

## Execution

```
burst-scheduler --burstNode virtual-kubelet-myaciconnector-linux --burstValue 10
```

Arguments:

| Argument | Type | Description |
|---|---|---|
| burstNode | String | Node name of the burst node. This is the node on which pods are scheduled once the burstValue has been met. |
| burstValue | Int | Value that controls how many pods will be scheduled on Kubernetes nodes vs. burst node. |
| kubeConfig | Bool | Indicates that a kubernetes config file found at $KUBECONFIG is used for cluster discovery / auth. If not specified, it is assumed execution is occurring from a pod in the Kubernetes cluster. |

## TODO:

**Terminating pods** – filter these from scope. Not a big issue but can be problematic during demos / quick turn-a-rounds.

**Namespace** - currently 'default' is a non-configurable default. Update with a `--namespace` argument.

**Default Scheduler** - Update pod updater to use default scheduler when not in burst. Currently a random node from all nodes - the burst node is chosen for scheduling. I am not able to patch the pod scheduler property value.

**Pod node assignment** - Update this to use go client method, go client rest interface. Currently the node assignment is handled through direct api call (via kubectl proxy / sidecar). I was unable to use podInterface.update due to a non-updatable property (error below). Is there a way to complete through the go client, if not I see the that there is a REST interface, this is probably a better way to achieve this.

```
"aci-helloworld-4142002832-3l873" is invalid: spec: Forbidden: pod updates may not change fields other than `spec.containers[*].image`, `spec.initContainers[*].image`, `spec.activeDeadlineSeconds` or `spec.tolerations` (only additions to existing tolerations)
```

https://github.com/kubernetes/kubernetes/issues/24913

## Incubation?

**Pseudo Re-scheduler** - could solve the problem with bursting a deployment. As the replica count reaches burst value (on pod delete), check pod balance. if pods are scheduled on a burst node, stop pod, which should then be rescheduled on a node via the custom scheduler - perhaps?