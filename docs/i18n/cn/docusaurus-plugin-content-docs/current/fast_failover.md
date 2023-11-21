---
sidebar_position: 9
sidebar_label: "故障恢复"
---

# 故障恢复

针对 Kubernetes 中的有状态应用（挂载了 HwameiStor PVC 的 Pod ），当 Pod 或者 PVC 出现问题时，尤其是 Kubernetes 节点出现问题时，
需要及时发现并重新调度，将 Pod 调度到其他健康的节点，并能成功挂载 PVC。
由于 Kubernetes 调度机制的限制，需要先等待比较长的时间（e.g. 5分钟）才能确定可以重新调度 Pod。
此外，由于 Pod 挂载了 PVC，还需额外等待较长时间（e.g. 6分钟）。
如果是 Statefulset 的 Pod，Kubernetes 不会进行重新调度，Deployment 的 Pod 可以。
这种情况将导致应用中断比较长时间，无法继续正常提供业务。

HwameiStor 为解决这类故障，提供了应用故障快速快速的能力。
在发现应用出现故障时，在很短的时间内将应用调度至另外的健康节点，同时保证在新节点上有应用所需的数据卷副本，从而保证业务应用正常运行。

## 使用方式

HwameiStor 为两类情况提供了应用故障快速恢复机制：

* 节点出现故障
  
  在这种情况下，该节点上的应用均无法正常运行。对于使用 HwameiStor 数据卷的应用，需要及时地将 Pod 重新调度到新的健康节点。
  您可以通过下列方式进行故障恢复：

  为该节点打标签（Label）：

  ```bash
  kubectl label node <nodeName> hwameistor.io/failover=start
  ```

  当故障恢复完成后，上面的标签会变成：

  ```console
  hwameistor.io/failover=completed
  ```
  
* 应用 Pod 出现故障

  在这种情况下，您可以为该 Pod 打标签（Label）对 Pod 进行故障恢复：

  ```bash
  kubectl label pod <podName> hwameistor.io/failover=start
  ```

  当故障恢复完成后，旧的 Pod 会被删除，新的 Pod 会在新的节点上启动并正常运行。
