---
sidebar_position: 2
sidebar_label: "CAS"
---

# CAS

Container Attached Storage (CAS) is software that includes microservice based storage
controllers that are orchestrated by Kubernetes. These storage controllers can run
anywhere that Kubernetes can run which means any cloud or even bare metal server or
on top of a traditional shared storage system. Critically, the data itself is also
accessed via containers as opposed to being stored in an off platform shared scale
out storage system. Because CAS leverages a microservices architecture, it keeps the
storage solution closely tied to the application bound to the physical storage device,
reducing I/O latency.

CAS is a pattern very much in line with the trend towards disaggregated data and the
rise of small, autonomous teams running small, loosely coupled workloads. For example,
my team might need Postgres for our microservice, and yours might depend on Redis and
MongoDB. Some of our use cases might require performance, some might be gone in 20 minutes,
others are write intensive, others read intensive, and so on. In a large organization,
the technology that teams depend on will vary more and more as the size of the organization
grows and as organizations increasingly trust teams to select their own tools.

CAS means that developers can work without worrying about the underlying requirements of
their organizations' storage architecture. To CAS, a cloud disk is the same as a SAN which
is the same as bare metal or virtualized hosts. Developers and Platform SREs don’t have
meetings to select the next storage vendor or to argue for settings to support their use
case, instead Developers remain autonomous and can spin up their own CAS containers with
whatever storage is available to the Kubernetes clusters.

CAS reflects a broader trend of solutions – many of which are now part of Cloud Native
Foundation – that reinvent particular categories or create new ones – by being built on
Kubernetes and microservice and that deliver capabilities to Kubernetes based microservice
environments. For example, new projects for security, DNS, networking, network policy
management, messaging, tracing, logging and more have emerged in the cloud-native ecosystem
and often in CNCF itself.

## Advantages of CAS

### Agility

Each storage volume in CAS has a containerized storage controller and corresponding
containerized replicas. Hence, maintenance and tuning of the resources around these
components are truly agile. The capability of Kubernetes for rolling upgrades enables
seamless upgrades of storage controllers and storage replicas. Resources such as CPU
and memory can be tuned using container cgroups.

### Granularity of Storage Policies

Containerizing the storage software and dedicating the storage controller to each
volume brings maximum granularity in storage policies. With CAS architecture, you
can configure all storage policies on a per-volume basis. In addition, you can
monitor storage parameters of every volume and dynamically update storage policies
to achieve the desired result for each workload. The control of storage throughput,
IOPS, and latency increases with this additional level of granularity in the volume
storage policies.

### Avoid Lock-in

Avoiding cloud vendor lock-in is a common goal for many Kubernetes users. However,
the data of stateful applications often remains dependent on the cloud provider and
technology or on an underlying traditional shared storage system, NAS or SAN. With
the CAS approach, storage controllers can migrate the data in the background per
workload and live migration becomes simpler. In other words, the granularity of
control of CAS simplifies the movement of stateful workloads from one Kubernetes
cluster to another in a non-disruptive way.

### Cloud Native

CAS containerizes the storage software and uses Kubernetes Custom Resource Definitions (CRDs)
to represent low-level storage resources, such as disks and storage pools. This model enables
storage to be integrated into other cloud-native tools seamlessly. The storage resources can
be provisioned, monitored, and managed using cloud-native tools such as Prometheus, Grafana,
Fluentd, Weavescope, Jaeger, and others.

Similar to hyperconverged systems, storage and performance of a volume in CAS are scalable.
As each volume has it's own storage controller, the storage can scale up within the permissible
limits of a storage capacity of a node. As the number of container applications increases in
a given Kubernetes cluster, more nodes are added, which increases the overall availability of
storage capacity and performance, thereby making the storage available to the new application containers.

### Lower Blast Radius

Because the CAS architecture is per workload and components are loosely coupled, CAS has a much
smaller blast radius than a typical distributed storage architecture.

CAS can deliver high availability through synchronous replication from storage controllers to
storage replicas. The metadata required to maintain the replicas is simplified to store the
information of the nodes that have replicas and information about the status of replicas to
help with quorum. If a node fails, the storage controller, which is a stateless container in
this case, is spun on a node where second or third replica is running and data continues to be
available. Hence, with CAS the blast radius is much lower and also localized to the volumes that
have replicas on that node.
