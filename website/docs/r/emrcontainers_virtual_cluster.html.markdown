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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `container_provider` - (Required) Configuration block for the container provider associated with your cluster.
* `name` - (Required) Name of the virtual cluster.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### container_provider Arguments

* `id` - The name of the container provider that is running your EMR Containers cluster
* `info` - Nested list containing information about the configuration of the container provider
    * `eks_info` - Nested list containing EKS-specific information about the cluster where the EMR Containers cluster is running
        * `namespace` - The namespace where the EMR Containers cluster is running
* `type` - The type of the container provider

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cluster.
* `id` - The ID of the cluster.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EKS Clusters using the `id`. For example:

```terraform
import {
  to = aws_emrcontainers_virtual_cluster.example
  id = "a1b2c3d4e5f6g7h8i9j10k11l"
}
```

Using `terraform import`, import EKS Clusters using the `id`. For example:

```console
% terraform import aws_emrcontainers_virtual_cluster.example a1b2c3d4e5f6g7h8i9j10k11l
```
