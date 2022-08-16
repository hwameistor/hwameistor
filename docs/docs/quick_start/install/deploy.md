---
sidebar_position: 2
sidebar_label: "Deploy by Helm Charts"
---

# Deploy by Helm Charts

The entire HwameiStor stack can be easily deployed by Helm Charts.

## Steps

### 1. Prepare helm tool

To install [Helm](https://helm.sh/) commandline tool, please refer to [Helm's Documentation](https://helm.sh/docs/).

### 2. Download `hwameistor` repo 

Download and extract repo file to the local directory.

```console
$ helm repo add hwameistor http://hwameistor.io/hwameistor

$ helm repo update hwameistor

$ helm pull hwameistor/hwameistor --untar
```

### 3. Deploy HwameiStor

```console
$ helm install hwameistor ./hwameistor \
    -n hwameistor --create-namespace
```

*That's it!*

To verify the deployment, please refer to the next chapter [Post Deployment](./post_check.md)

## Use image repository mirrors

:::tip

The default image repositories are `quay.io` and `ghcr.io`.
In case they are blocked in some places, DaoCloud provides their mirrors at `quay.m.daocloud.io` and `ghcr.m.daocloud.io`

:::

To switch image repository mirrors, use `--set` to change the value of parameters: `k8sImageRegistry` and `hwameistorImageRegistry`

```console
$ helm install hwameistor ./hwameistor \
    -n hwameistor --create-namespace \
    --set k8sImageRegistry=quay.m.daocloud.io \
    --set hwameistorImageRegistry=ghcr.m.daocloud.io
```

## Customize kubelet root directory

:::caution

The default `kubelet` directory is `/var/lib/kubelet`.
If your Kubernetes distribution uses a different `kubelet` directory, you must set the parameter `kubeletRootDir`.

:::

For example, on [Canonical's MicroK8s](https://microk8s.io/) which uses `/var/snap/microk8s/common/var/lib/kubelet/` as `kubelet` directory,  HwameiStor needs to be deployed as:
 
```console
$ helm install hwameistor ./hwameistor \
    -n hwameistor --create-namespace \
    --set kubeletRootDir=/var/snap/microk8s/common/var/lib/kubelet/
```

## Production setup

A production environment would require:

- specify resource configuration
- avoid deploying on master nodes
- implement quick failover of controllers

We provide some recommended values in `values.extra.prod.yaml`, to use it:

```console
$ helm install hwameistor ./hwameistor \
    -n hwameistor --create-namespace \
    -f ./hwameistor/values.yaml \
    -f ./hwameistor/values.extra.prod.yaml
```

:::caution

In a resource-strained test environment, setting the above-mentioned values would cause pods unable to start! 
:::
