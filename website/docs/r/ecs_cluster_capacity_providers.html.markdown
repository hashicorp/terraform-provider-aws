---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_cluster_capacity_providers"
description: |-
  Provides an ECS cluster capacity providers resource.
---

# Resource: aws_ecs_cluster_capacity_providers

Manages the capacity providers of an ECS Cluster.

More information about capacity providers can be found in the [ECS User Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cluster-capacity-providers.html).

~> **NOTE on Clusters and Cluster Capacity Providers:** Terraform provides both a standalone `aws_ecs_cluster_capacity_providers` resource, as well as allowing the capacity providers and default strategies to be managed in-line by the [`aws_ecs_cluster`](/docs/providers/aws/r/ecs_cluster.html) resource. You cannot use a Cluster with in-line capacity providers in conjunction with the Capacity Providers resource, nor use more than one Capacity Providers resource with a single Cluster, as doing so will cause a conflict and will lead to mutual overwrites.

## Example Usage

```terraform
resource "aws_ecs_cluster" "example" {
  name = "my-cluster"
}

resource "aws_ecs_cluster_capacity_providers" "example" {
  cluster_name = aws_ecs_cluster.example.name

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}
```

## Argument Reference

The following arguments are supported:

* `capacity_providers` - (Optional) Set of names of one or more capacity providers to associate with the cluster. Valid values also include `FARGATE` and `FARGATE_SPOT`.
* `cluster_name` - (Required, Forces new resource) Name of the ECS cluster to manage capacity providers for.
* `default_capacity_provider_strategy` - (Optional) Set of capacity provider strategies to use by default for the cluster. Detailed below.

### default_capacity_provider_strategy Configuration Block

* `capacity_provider` - (Required) Name of the capacity provider.
* `weight` - (Optional) The relative percentage of the total number of launched tasks that should use the specified capacity provider. The `weight` value is taken into consideration after the `base` count of tasks has been satisfied. Defaults to `0`.
* `base` - (Optional) The number of tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined. Defaults to `0`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Same as `cluster_name`.

## Import

ECS cluster capacity providers can be imported using the `cluster_name` attribute. For example:

```
$ terraform import aws_ecs_cluster_capacity_providers.example my-cluster
```
