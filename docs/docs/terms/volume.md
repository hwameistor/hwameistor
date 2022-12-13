---
sidebar_position: 5
sidebar_label: "Volume"
---

# Volume

On-disk files in a container are ephemeral, which presents some problems for non-trivial
applications when running in containers.

- One problem is the loss of files when a container crashes.
  The kubelet restarts the container but with a clean state.
- A second problem occurs when sharing files between containers running together in a `Pod`.

The Kubernetes volume abstraction solves both of these problems.

Kubernetes supports many types of volumes. A Pod can use any number of volume types simultaneously.
Ephemeral volume types have a lifetime of a pod, but persistent volumes exist beyond the lifetime
of a pod. When a pod ceases to exist, Kubernetes destroys ephemeral volumes; however, Kubernetes
does not destroy persistent volumes. For any kind of volume in a given pod, data is preserved
across container restarts.

At its core, a volume is a directory, possibly with some data in it, which is accessible to the
containers in a pod. How that directory comes to be, the medium that backs it, and the contents
of it are determined by the particular volume type used.

To use a volume, specify the volumes to provide for the Pod in `.spec.volumes` and declare where
to mount those volumes into containers in `.spec.containers[*].volumeMounts`.

See also the official documentation provided by Kubernetes:

- [Volume](https://kubernetes.io/docs/concepts/storage/volumes/)
- [Persistent Volume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Ephemeral Volume](https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/)
