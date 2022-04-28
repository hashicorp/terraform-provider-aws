---
subcategory: "EMR Containers"
layout: "aws"
page_title: "AWS: aws_emrcontainers_virtual_cluster"
description: |-
  Manages an EMR Containers (EMR on EKS) Virtual Cluster
---

# Resource: aws_emrcontainers_virtual_cluster

Manages an EMR Containers (EMR on EKS) Virtual Cluster.

## Example Usage

### Basic Usage

```terraform
resource "aws_emrcontainers_virtual_cluster" "example" {
  container_provider {
    id   = aws_eks_cluster.example.name
    type = "EKS"

    info {
      eks_info {
        namespace = "default"
      }
    }
  }

  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `container_provider` - (Required) Configuration block for the container provider associated with your cluster.
* `name` – (Required) Name of the virtual cluster.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### container_provider Arguments

* `id` - The name of the container provider that is running your EMR Containers cluster
* `info` - Nested list containing information about the configuration of the container provider
    * `eks_info` - Nested list containing EKS-specific information about the cluster where the EMR Containers cluster is running
        * `namespace` - The namespace where the EMR Containers cluster is running
* `type` - The type of the container provider

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the cluster.
* `id` - The ID of the cluster.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

EKS Clusters can be imported using the `name`, e.g.

```
$ terraform import aws_emrcontainers_virtual_cluster.example
```
