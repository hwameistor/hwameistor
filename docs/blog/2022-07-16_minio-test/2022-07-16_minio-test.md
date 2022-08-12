---
slug: minio
title: HwameiStor Supports MinIO
authors: [Simon, Michelle]
tags: [Test]
---

# HwameiStor Supports MinIO

This blog introduces an MinIO storage solution built on HwameiStor, and clarifies the detailed test procedures about whether HwameiStor can properly support those basic features and tenant isolation function provided by MinIO.

## MinIO introduction

MinIO is a high performance object storage solution with native support for Kubernetes deployments.
It can provide distributed, S3-compatible, and multi-cloud storage service in public cloud, private cloud,
and edge computing scenarios. MinIO is a software-defined product and released under [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.en.html).
It can also run well on x86 and other standard hardware.

![MinIO design](minio-design.png)

MinIO is designed to meet private cloud's requirements for high performance,
in addition to all required features of object storage.
MinIO features easy to use, cost-effective, and high performance in providing scalable cloud-native object storage services.

MinIO works well in traditional object storage scenarios, such as secondary storage, disaster recovery, and archiving.
It also shows competitive capabilities in machine learning, big data, private cloud, hybrid cloud,
and other emerging fields to well support data analysis, high performance workloads, and cloud-native applications.

### MinIO architecture

MinIO is designed for the cloud-native architecture, so it can be run as a lightweight container
and managed by external orchestration tools like Kubernetes.

The MinIO package comprises of static binary files less than 100 MB.
This small package enables it to efficiently use CPU and memory resources even
with high workloads and can host a large number of tenants on shared hardware.

MinIO's architecture is as follows:

![Architecture](architect.png)

MinIO can run on a standard server that have installed proper local drivers (JBOD/JBOF).
An MinIO cluster has a totally symmetric architecture. In other words,
each server provide same functions, without any name node or metadata server.

MinIO can write both data and metadata as objects, so there is no need to use metadata servers.
MinIO provides erasure coding, bitrot protection, encryption and other features in a strict and consistent way.

Each MinIO cluster is a set of distributed MinIO servers, one MinIO process running on each node.

MinIO runs in a userspace as a single process, and it uses lightweight co-routines for high concurrence.
It divides drivers into erasure sets (generally 16 drivers in each set),
and uses the deterministic hash algorithm to place objects into these erasure sets.

MinIO is specifically designed for large-scale and multi-datacenter cloud storage service.
Tenants can run their own MinIO clusters separately from others, getting rid of interruptions
from upgrade or security problems. Tenants can scale up by connecting multi clusters across geographical regions.

![node-distribution-setup](node-setup.png)

## Build test environment

### Deploy Kubernetes cluster

A Kubernetes cluster was deployed with three virtual machines: one as the master node and two as worker nodes. The kubelet version is 1.22.0.

![k8s-cluster](k8s-cluster.png)

### Deploy HwameiStor local storage

Deploy HwameiStor local storage on Kubernetes:

![check HwameiStor local storage](kubectl-get-hwamei-pod.png)

Allocate five disks (SDB, SDC, SDD, SDE, and SDF) for each worker node to support HwameiStor local disk management:

![lsblk](lsblk01.png)

![lsblk](lsblk02.png)

Check node status of local storage:

![get-lsn](kubectl-get-lsn.png)

Create storageClass:

![get-sc](kubectl-get-sc.png)

## Deploy distributed multi-tenant cluster (minio-operator)

This section will show how to deploy minio-operator, how to create a tenant,
and how to configure HwameiStor local volumes.

### Deploy minio-operator

1. Copy minio-operator repo to your local environment

  ```
  git clone <https://github.com/minio/operator.git>
  ```

  ![helm-repo-list](helm-repo-list.png)

  ![ls-operator](ls-opeartor.png)

2. Enter helm operator directory `/root/operator/helm/operator`

  ![ls-pwd](ls-pwd.png)

3. Deploy the minio-operator instance

  ```
  helm install minio-operator \
  --namespace minio-operator \
  --create-namespace \
  --generate-name .
  --set persistence.storageClass=local-storage-hdd-lvm .
  ```

4. Check minio-operator running status

  ![get-all](kubectl-get-all.png)

### Create tenants


1. Enter the `/root/operator/examples/kustomization/base` directory and change `tenant.yaml`

  ![git-diff-yaml](git-diff-tenant-yaml.png)

2. Enter the `/root/operator/helm/tenant/` directory and change `values.yaml`

  ![git-diff-values.yaml](git-diff-values-yaml.png)

3. Enter `/root/operator/examples/kustomization/tenant-lite` directory and change `kustomization.yaml`

  ![git-diff-kustomization-yaml](git-diff-kustomization-yaml.png)

4. Change `tenant.yaml`

  ![git-diff-tenant-yaml02](git-diff-tenant-yaml02.png)

5. Change `tenantNamePatch.yaml`

  ![git-diff-tenant-name-patch-yaml](git-diff-tenant-name-patch-yaml.png)

6. Create a tenant

  ```
  kubectl apply –k . 
  ```

7. Check resource status of the tenant minio-t1

  ![kubectl-get-all-nminio-tenant](kubectl-get-all-nminio-tenant.png)

8. To create another new tenant, you can first create a new directory `tenant` (in this example `tenant-lite-2`) under `/root/operator/examples/kustomization` and change the files listed above

  ![pwd-ls-ls](pwd-ls-ls.png)

9. Run `kubectl apply –k .` to create the new tenant `minio-t2`

  ![kubectl-get-all-nminio](kubectl-get-all-minio.png)

### Configure HwameiStor local volumes

Run the following commands in sequence to finish this configuration:

```
kubectl get statefulset.apps/minio-t1-pool-0 -nminio-tenant -oyaml
```

![local-storage-hdd-lvm](local-storage-hdd-lvm.png)

```
kubectl get pvc –A
```

![kubectl-get-pvc](kubectl-get-pvc.png)

```
kubectl get pvc export-minio6-0 -nminio-6 -oyaml
```

![kubectl-get-pvc-export-oyaml](kubectl-get-pvc-export-oyaml.png)

```
kubectl get pv
```

![kubectl-get-pv](kubectl-get-pv.png)

```
kubectl get pvc data0-minio-t1-pool-0-0 -nminio-tenant -oyaml
```

![kubectl-get-pvc-oyaml](kubectl-get-pvc-oyaml.png)

```
kubectl get lv
```

![kubectl-get-lv](kubectl-get-lv.png)

```
kubect get lvr
```

![kubectl-get-lvr](kubectl-get-lvr.png)

## Test HwameiStor's support for MinIo

With the above settings in place, now let's test basic features and tenant isolation.

### Test basic features


1. Log in to `minio console：10.6.163.52:30401/login`

  ![minio-opeartor-console-login](minio-opeartor-console-login.png)

2. Get JWT by `kubectl minio proxy -n minio-operator`

  ![minio-opeartor-console-login](kubectl-minio-proxy-jwt.png)

3. Browse and manage information about newly-created tenants

  ![tenant01](tenant01.png)

  ![tenant02](tenant02.png)

  ![tenant03](tenant03.png)

  ![tenant04](tenant04.png)

  ![tenant05](tenant05.png)

  ![tenant06](tenant06.png)

4. Log in as tenant minio-t1 (Account: minio)

  ![login-minio](login-minio-t1-01.png)

  ![login-minio](login-minio-t1-02.png)

5. Browse bucket bk-1

  ![view-bucket-1](view-bucket-01.png)

  ![view-bucket-1](view-bucket-02.png)

  ![view-bucket-1](view-bucket-03.png)

6. Create a new bucket bk-1-1

  ![create-bucket-1-1](create-bucket-1-1.png)

  ![create-bucket-1-1](create-bucket-1-2.png)

  ![create-bucket-1-1](create-bucket-1-3.png)

7. Create path path-1-2

  ![create-path-1-2](create-path-1-2-01.png)

  ![create-path-1-2](create-path-1-2-02.png)

8. Upload the file

  ![upload-file](upload-file-success.png)

  ![upload-file](upload-file-success-02.png)

  ![upload-file](upload-file-success-03.png)

9. Upload the folder

  ![upload-folder](upload-folder-success-01.png)

  ![upload-folder](upload-folder-success-02.png)

  ![upload-folder](upload-folder-success-03.png)

  ![upload-folder](upload-folder-success-04.png)

10. Create a user with read-only permission

  ![create-user](create-readonly-user-01.png)

  ![create-user](create-readonly-user-02.png)

### Test tenant isolation

1. Log in as tenant minio-t2

  ![login-t2](login-minio-t2-01.png)

  ![login-t2](login-minio-t2-02.png)

2. Only minio-t2 information is visible. You cannot see information about tenant minio-t1.

  ![only-t2](only-t2.png)

3. Create bucket

  ![create-bucket](create-bucket01.png)

  ![create-bucket](createbucket02.png)

4. Create path

  ![create-path](create-path01.png)

  ![create-path](create-path02.png)

5. Upload the file

  ![upload-file](upload-file01.png)

  ![upload-file](upload-file02.png)

6. Create a user

  ![create-user](create-user01.png)

  ![create-user](create-user02.png)

  ![create-user](create-user03.png)

  ![create-user](create-user04.png)

  ![create-user](create-user05.png)

7. Configure user policies

  ![user-policy](user-policy01.png)

  ![user-policy](user-policy02.png)

8. Delete a bucket

  ![delete-bucket](delete-bk01.png)

  ![delete-bucket](delete-bk02.png)

  ![delete-bucket](delete-bk03.png)

  ![delete-bucket](delete-bk04.png)

  ![delete-bucket](delete-bk05.png)

  ![delete-bucket](delete-bk06.png)

## Conclusion

In this test, we successfully deployed MinIO distributed object storage on the basis of Kubernetes 1.22 and
the HwameiStor local storage. We performed the basic feature test,
system security test, and operation and maintenance management test.

All tests are passed, proving HwameiStor can well support for MinIO.
