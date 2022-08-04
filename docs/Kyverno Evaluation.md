# `Kyverno` Evaluation

This document evaluates [`Kyverno`](https://`Kyverno`.io/) as a replacement to [`Kupid`](https://github.com/gardener/`Kupid`) for long term usage and sustainability. 

## Goal
The goal of this document is to provide - 
* Worthiness of `Kyverno` as an elegant/sustainable replacement for `Kupid`
* Results of realizing an existing `Kupid` based ClusterPolicy using similar policy from `Kyverno` and testing it on a local seed.
* Obvious gaps which needs consideration before adoption
* A glimpse on usage of `Kyverno` beyond the scope of current `Kupid` usage and choices it offers. 
* Cursory comparison to alternatives like  [opa-gatekeeper](https://open-policy-agent.github.io/gatekeeper/website/docs/). 

## Out Of Scope 
The document does not intend to provide an exhaustive study on `Kyverno` or other alternatives to `Kupid`. 
It does aims to set the stage for a deeper study and scoping of work if we see `Kyverno` as a good fit from this evaluation.


## `Kyverno` Features
- Like `Kupid`, `Kyverno` is declarative and creates policies as k8s resource. 
- Like `Kupid`, it supports namespaced and cluster wide policy definitions. 
- Like `Kupid`, it supports validate and mutate operations. 
- Unlike `Kupid`, it can be used to generate resources or to verify images.
- Unlike `Kupid`, it provides monitoring and reporting capabilities which can ease the operators life. 
- Unlike `Kupid`, it provides more exhaustive usage of matchers and selectors:
  - Can have Match or optional exclude declarations. 
  - Has support for resource filters which can be specified with any or all clause - 
     
    - **resources**: select resources by names, namespaces, kinds, label selectors, annotations, and namespace selectors.
    - **subjects**: select users, user groups, and service accounts
    - **roles**: select namespaced roles
    - **clusterRoles**: select cluster wide roles
-  A mutate rule can be used to modify matching resources and is written as either a [RFC 6902 JSON Patch](http://jsonpatch.com/) or a [strategic merge patch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md).
- Unlike `Kupid`, With `Kyverno` 1.7.0+, `Kyverno` supports the mutate existing resources with patchesStrategicMerge and patchesJson6902. 
  - Mutate existing policies are applied in the background which updates existing resources in the cluster

As highlighted in the out-of-scope section earlier, this document is not intended to be a replacement for the official documentation of `Kyverno`. 
However, for completeness. it's worth mentioning features highlighed below which allows more use cases to be supported for injecting policies in Gardener.
- [Validate Resources](https://`Kyverno`.io/docs/writing-policies/validate/)
- [Verify Images](https://`Kyverno`.io/docs/writing-policies/verify-images/)
- [Generate Resources](https://`Kyverno`.io/docs/writing-policies/generate/)
- [Variables](https://`Kyverno`.io/docs/writing-policies/variables/)
- [External Data Sources](https://`Kyverno`.io/docs/writing-policies/external-data-sources/)
- [Preconditions](https://`Kyverno`.io/docs/writing-policies/preconditions/)
- [Auto-Gen rules for Pod controllers](https://`Kyverno`.io/docs/writing-policies/autogen/)
- [Background scans](https://`Kyverno`.io/docs/writing-policies/background/)

**Last but not least it is supported by community, so it removes maintenance overhead currently faced with `Kupid`.** 

## Trying out `Kyverno` 

We tried to use `Kyverno` to adapt an existing `Kupid` based `ClusterPodSchedulingPolicy` with and equivalent `Kyverno` `ClusterPolicy`. 

**The cluster policy defined using `Kyverno` for the corresponding `Kupid` ClusterPodSchedulingPolicy**
ClusterPolicy for Dedicated Worker Pool 

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: add-etcd-node-affinity
  annotations:
    policies.kyverno.io/title: Add Node Affinity for etcd stateful sets
    policies.kyverno.io/category: Other
    policies.kyverno.io/severity: medium
    policies.kyverno.io/subject: Deployment
    kyverno.io/kyverno-version: 1.7.0
    kyverno.io/kubernetes-version: "1.23"
    policies.kyverno.io/description: >-
      Node affinity, similar to node selection, is a way to specify which node(s) on which Pods will be scheduled
      but based on more complex conditions. This policy will add node affinity for the statefulsets of etcd-main and loki
      to be deployed to a dedicated worker pool.
spec:
  background: false
  rules:
  - name: add-etcd-node-affinity
    match:
      resources:
        kinds:
        - apps/v1/StatefulSet
        names:
        - etcd-main
        - loki
    mutate:
      patchesJson6902: |-
        - path: "/spec/template/spec/affinity/nodeAffinity/requiredDuringSchedulingIgnoredDuringExecution/nodeSelectorTerms/-1/matchExpressions/-1"
          op: add
          value: {"key":"pool.worker.gardener.cloud/dedicated-for","operator":"In","values":["etcd"]}
        - path: "/spec/template/spec/tolerations/-1"
          op: add
          value: {"effect":"NoExecute","key":"pool.worker.gardener.cloud/dedicated-for","value":"etcd"}
```

### Verifying the results post application of policy 

<details><summary> After applying the policy we could see that the etcd-main and loki sts were scheduled on the same node while etcd-events is not. </summary>

  ```sh
  $> kgp -o wide | grep -E etcd\|loki
  etcd-events-0                                 2/2     Running   0              22d    100.64.1.139   ip-10-180-16-211.eu-west-1.compute.internal   <none>           <none>
  etcd-main-0                                   2/2     Running   0              13h    100.64.0.22    ip-10-180-2-85.eu-west-1.compute.internal     <none>           <none>
  loki-0                                        4/4     Running   0              13h    100.64.0.21    ip-10-180-2-85.eu-west-1.compute.internal     <none>           <none>
  ```
</details>

<details> <summary>Checking the labels for the node  ip-10-180-2-85.eu-west-1.compute.internal it does show it to be a dedicated worker pool  </summary>

   ```yaml
  ❯ k describe node ip-10-180-2-85.eu-west-1.compute.internal
  Name:               ip-10-180-2-85.eu-west-1.compute.internal
  Roles:              <none>
  Labels:             beta.kubernetes.io/arch=amd64
                    beta.kubernetes.io/instance-type=m5.2xlarge
                    beta.kubernetes.io/os=linux
                    failure-domain.beta.kubernetes.io/region=eu-west-1
                    failure-domain.beta.kubernetes.io/zone=eu-west-1b
                    kubernetes.io/arch=amd64
                    kubernetes.io/hostname=ip-10-180-2-85.eu-west-1.compute.internal
                    kubernetes.io/os=linux
                    networking.gardener.cloud/node-local-dns-enabled=false
                    node.kubernetes.io/instance-type=m5.2xlarge
                    node.kubernetes.io/role=node
                    pool.worker.gardener.cloud/dedicated-for=etcd
                    topology.ebs.csi.aws.com/zone=eu-west-1b
                    topology.kubernetes.io/region=eu-west-1
                    topology.kubernetes.io/zone=eu-west-1b
                    worker.garden.sapcloud.io/group=ash-kyverno-etc
                    worker.gardener.cloud/cri-name=containerd
                    worker.gardener.cloud/kubernetes-version=1.23.4
                    worker.gardener.cloud/pool=ash-kyverno-etc
   ```
</details>

<details><summary>Checking the yaml of etcd sts show that `kyverno` has applied the desired annotation, affinities and tolerations </summary>

```yaml
    ❯ k get sts etcd-main -oyaml
      apiVersion: apps/v1
      kind: StatefulSet
      metadata:
        annotations:
        ..
        ..
          policies.kyverno.io/last-applied-patches: |
            add-etcd-node-affinity.add-etcd-node-affinity.kyverno.io: removed /spec/template/spec/tolerations/1
        creationTimestamp: "2022-06-09T07:32:47Z"
        labels:
          app: etcd-statefulset
          ..
          ..
          role: main
        name: etcd-main
        namespace: shoot--dev--ashlclkyv
      spec:
        replicas: 1
        selector:
          matchLabels:
            instance: etcd-main
            name: etcd
        template:
          metadata:
            annotations:
            ..
            creationTimestamp: null
            labels:
            ..
            ..
          spec:
            affinity:
              nodeAffinity:
                requiredDuringSchedulingIgnoredDuringExecution:
                  nodeSelectorTerms:
                  - matchExpressions:
                    - key: pool.worker.gardener.cloud/dedicated-for
                      operator: In
                      values:
                      - etcd
             ..
             ..
            containers:
            - 
              ..
              ...
            - 
               .. 
               ... 
            restartPolicy: Always
            schedulerName: default-scheduler
            securityContext: {}
            .. 
            ...
            tolerations:
            - effect: NoExecute
              key: pool.worker.gardener.cloud/dedicated-for
              operator: Equal
              value: etcd
            ..
            ...
            ....
          status:
            .. 
      status:
        availableReplicas: 1
        ..
        ...
```
</details>

**In conclusion `Kyverno` was able to achieve the same injection of policy as is currently implemented via `Kupid`**

**Some interesting noteworthy observations during the limited usage:**
- During our experiments, the background scanning feature seems to not work for Mutating policy above. We had do apply a dummy modification on the sts for the updated policy to be applied. [Although the documentation mentions its possible with 1.7+ version]
- Deleting the policy also doesn't remove the policies applied and needs to managed manually or as a delete rule before removing the policy. 
- One nice thing observed was the `kyverno` annotated the resource(sts) with the last operation applied by `kyverno`  - 
   ```yaml
      policies.kyverno.io/last-applied-patches: |
      add-etcd-node-affinity.add-etcd-node-affinity.kyverno.io: added /spec/template/spec/tolerations/1
   ```
- <details><summary> Also logs in `Kyverno`not only capture the operation applied but also log the patch applied as base64 encoded. </summary>

  ```
  I0804 14:29:26.060014       1 handlers.go:254] webhooks/resource/mutate "msg"="completed mutating webhook" "kind"={"group":"apps","version":"v1","kind":"StatefulSet"} "name"="loki" "namespace"="shoot--dev--ashlclkyv" "operation"="UPDATE" "uid"="8abbd720-4ecf-4f02-800b-c244fdb84deb" "response"={"uid":"","allowed":true,"patch":"W3sib3AiOiJhZGQiLCJwYXRoIjoiL3NwZWMvdGVtcGxhdGUvc3BlYy90b2xlcmF0aW9ucyIsInZhbHVlIjpbeyJlZmZlY3QiOiJOb0V4ZWN1dGUiLCJrZXkiOiJwb29sLndvcmtlci5nYXJkZW5lci5jbG91ZC9kZWRpY2F0ZWQtZm9yIiwidmFsdWUiOiJldGNkIn1dfSwgeyJvcCI6ImFkZCIsInBhdGgiOiIvc3BlYy90ZW1wbGF0ZS9zcGVjL2FmZmluaXR5IiwidmFsdWUiOnsibm9kZUFmZmluaXR5Ijp7InJlcXVpcmVkRHVyaW5nU2NoZWR1bGluZ0lnbm9yZWREdXJpbmdFeGVjdXRpb24iOnsibm9kZVNlbGVjdG9yVGVybXMiOlt7Im1hdGNoRXhwcmVzc2lvbnMiOlt7ImtleSI6InBvb2wud29ya2VyLmdhcmRlbmVyLmNsb3VkL2RlZGljYXRlZC1mb3IiLCJvcGVyYXRvciI6IkluIiwidmFsdWVzIjpbImV0Y2QiXX1dfV19fX19LCB7InBhdGgiOiIvbWV0YWRhdGEvYW5ub3RhdGlvbnMvcG9saWNpZXMua3l2ZXJuby5pb34xbGFzdC1hcHBsaWVkLXBhdGNoZXMiLCJvcCI6InJlcGxhY2UiLCJ2YWx1ZSI6ImFkZC1ldGNkLW5vZGUtYWZmaW5pdHkuYWRkLWV0Y2Qtbm9kZS1hZmZpbml0eS5reXZlcm5vLmlvOiBhZGRlZCAvc3BlYy90ZW1wbGF0ZS9zcGVjL2FmZmluaXR5XG4ifV0="}
  I0804 14:29:26.060673       1 admission.go:88] webhooks/resource/mutate "msg"="admission review request processed" "kind"={"group":"apps","version":"v1","kind":"StatefulSet"} "name"="loki" "namespace"="shoot--dev--ashlclkyv" "operation"="UPDATE" "uid"="8abbd720-4ecf-4f02-800b-c244fdb84deb" "time"="138.859546ms"

  ---------------------
  # Decoding the patch helps us easily trace back the patch applied
  $> echo W3sib3AiOiJhZGQiLCJwYXRoIjoiL3NwZWMvdGVtcGxhdGUvc3BlYy90b2xlcmF0aW9ucyIsInZhbHVlIjpbeyJlZmZlY3QiOiJOb0V4ZWN1dGUiLCJrZXkiOiJwb29sLndvcmtlci5nYXJkZW5lci5jbG91ZC9kZWRpY2F0ZWQtZm9yIiwidmFsdWUiOiJldGNkIn1dfSwgeyJvcCI6ImFkZCIsInBhdGgiOiIvc3BlYy90ZW1wbGF0ZS9zcGVjL2FmZmluaXR5IiwidmFsdWUiOnsibm9kZUFmZmluaXR5Ijp7InJlcXVpcmVkRHVyaW5nU2NoZWR1bGluZ0lnbm9yZWREdXJpbmdFeGVjdXRpb24iOnsibm9kZVNlbGVjdG9yVGVybXMiOlt7Im1hdGNoRXhwcmVzc2lvbnMiOlt7ImtleSI6InBvb2wud29ya2VyLmdhcmRlbmVyLmNsb3VkL2RlZGljYXRlZC1mb3IiLCJvcGVyYXRvciI6IkluIiwidmFsdWVzIjpbImV0Y2QiXX1dfV19fX19LCB7InBhdGgiOiIvbWV0YWRhdGEvYW5ub3RhdGlvbnMvcG9saWNpZXMua3l2ZXJuby5pb34xbGFzdC1hcHBsaWVkLXBhdGNoZXMiLCJvcCI6InJlcGxhY2UiLCJ2YWx1ZSI6ImFkZC1ldGNkLW5vZGUtYWZmaW5pdHkuYWRkLWV0Y2Qtbm9kZS1hZmZpbml0eS5reXZlcm5vLmlvOiBhZGRlZCAvc3BlYy90ZW1wbGF0ZS9zcGVjL2FmZmluaXR5XG4ifV0= | base64 -d
  [{"op":"add","path":"/spec/template/spec/tolerations","value":[{"effect":"NoExecute","key":"pool.worker.gardener.cloud/dedicated-for","value":"etcd"}]}, {"op":"add","path":"/spec/template/spec/affinity","value":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"pool.worker.gardener.cloud/dedicated-for","operator":"In","values":["etcd"]}]}]}}}}, {"path":"/metadata/annotations/policies.kyverno.io~1last-applied-patches","op":"replace","value":"add-etcd-node-affinity.add-etcd-node-affinity.kyverno.io: added /spec/template/spec/affinity\n"}]
  ```
</details>
There can be moer hidden features, but that can be considered in the wider scope of implementation. 

## Usage & Design Consideration
### `Kupid` replacement 
`Kupid` replacement will require following - 
- Replacing resources created via  `clusterpodschedulingpolicies.kupid.gardener.cloud` with corresponding resources using `clusterpolicies.kyverno.io`. 
- Deploying `Kyverno` in seeds which can be done via - 
  - Using the existing controllerRegistration/Deployment artifacts for `Kupid` and adapting them for `Kyverno`, as well as adapting the necessary changes in g/g botanist component model for `Kupid`.

     or 

  - Removing `Kupid` completely as an extension and instead deploying `Kyverno` in a fashion similar to [Managed Istio feature gate for Seed clusters](https://github.com/gardener/gardener/pull/2273).


> We should also explore the [Reporting](https://`Kyverno`.io/docs/policy-reports/) and [Monitoring](https://`Kyverno`.io/docs/monitoring/) capabilities of `Kyverno` and bring them as part of the replacement strategy to make the solution complete. 

 
## Beyond just a `Kupid` replacement 
As `Kyverno` feature set is quite rich, we can also consider it for other scenarios which can be of value. 

### Additional Usage 
Features which can be considered additionally are - 
- The `Kyverno` [CLI test command](https://`Kyverno`.io/docs/`Kyverno`-cli/#test) can be used to test if a policy is behaving as expected.
- Validate resources using the `Kyverno` CLI, in your CI/CD pipeline, before applying to your cluster
- Block non-conformant resources using admission controls, or report policy violations
- Validate and mutate using overlays (like `Kustomize!`)

### Operator Usage 
- Once we have `Kyverno` deployed in the landscape, we can also consider to consume it to manage all orthogonal changes which might be helpful for day-2-day operations of landscapes.
- We can have controllers like Cluster Health or Zonal Health to act by creating policies on the fly to manage the resources on a seed. 
- Can we used as an easy tooling for migrating worklod during outages. 
- Can also be considered for handling infra limitations/implications onDemand.

### Shoot Owner Usage 
Also though not thought through, we can think of enabling `kyverno` for consumers of Gardener as a functional differentiator to allow the Shoot owners to play with Shoot semantics and gardener provisioned resources underneath. 
This is just a wild thought and need to be considered with all pros and cons. 

## Alternatives
One of the well known alternatives to consider is  [opa-gatekeeper](https://open-policy-agent.github.io/gatekeeper/website/docs/). However we see some challenges with its usage when compared with `Kyverno` namely - 
- `opa-gatekeeper` is a more general purpose tool and can handle non k8s resources as well. However, it may not be relevance as in Gardener we primarily work with k8s resources.
- `opa-gatekeeper` also has a steeper learning curve given its imperative style and design overhead to define policy in an opinionated language. 
- If you are still not convinced watch this [video comparison](https://www.youtube.com/watch?v=9gSrRNmmKBc)

## Conclusion
In our analysis `Kyverno` can not only be a good candidate to replace `Kupid` but it can also be used to extend the functionality of orthogonal management of resources to a wider scope in Gardener. 