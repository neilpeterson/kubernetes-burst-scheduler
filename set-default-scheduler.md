# How to update scheduler to default scheduler

## Attribution

Most of these ideas orginated here - https://github.com/songbinliu/updatePod

## Abstract

The burst scheduler should schedule pods on non-bound / standard Kubernetes nodes, and only burst to the burst node once the burst value has been breached.

Currently, if the burst value has not been breached, a random node is selected for pod placement. The desired outcome is rather than a random node, the pod spec is updated to use the default scheduler. This however is problematic; the scheduler name is a non-updatable pod attribute.

## Thoughts on situation 1 – replica set 

If it can be proven that a pod has been created by a replica set, the replicas set template can be updated to have the default scheduler. The pod pending assignment is then deleted. A new pod is created by the replica set with the default scheduler in the spec.

Potentially messy and will need to track the state of the replica set. I see no other viable options.

## Thoughts on situation 2 – no replica set

Capture pod spec, delete pod, re-create with default scheduler.

Potentially messy. I see no other viable options.
