---
sidebar_position: 4
sidebar_label: "CRD and CR"
---

# CRD and CR

## CRD

`CRD` is the abbreviation of `Custom Resource Definition`, and is a resource type
natively provided by `Kubernetes`. It is the definition of Custom Resource (CR)
to describe what a custom resource is.

A CRD can register a new resource with the `Kubernetes` cluster to extend the
capabilities of the `Kubernetes` cluster. With `CRD`, you can define the abstraction
of the underlying infrastructure, customize resource types based on business needs,
and use the existing resources and capabilities of `Kubernetes` to define higher-level
abstractions through a Lego-like building blocks.

## CR

`CR` is the abbreviation of `Custom Resource`. In practice, it is an instance of `CRD`,
a resource description that matches with the field format in `CRD`.

## CRDs + Controllers

We all know that `Kubernetes` has powerful scalability, but only `CRD` is not useful.
It also needs the support of controller (`Custom Controller`) to reflect the value of `CRD`.
`Custom Controller` can listen `CRUD` events of `CR` to implement custom business logic.

In `Kubernetes`, `CRDs + Controllers = Everything`.

See also the official documentation provided by Kubernetes:

- [CustomResource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [CustomResourceDefinition](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
