---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_cluster_capacity_providers"
description: |-
  Provides an ECS cluster capacity providers resource.
---

# Resource: aws_ecs_cluster_capacity_providers

Manages the capacity providers of an ECS Cluster.

More information about capacity providers can be found in the [ECS User Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cluster-capacity-providers.html).

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

This resource supports the following arguments:

* `capacity_providers` - (Optional) Set of names of one or more capacity providers to associate with the cluster. Valid values also include `FARGATE` and `FARGATE_SPOT`.
* `cluster_name` - (Required, Forces new resource) Name of the ECS cluster to manage capacity providers for.
* `default_capacity_provider_strategy` - (Optional) Set of capacity provider strategies to use by default for the cluster. Detailed below.

### default_capacity_provider_strategy Configuration Block

* `capacity_provider` - (Required) Name of the capacity provider.
* `weight` - (Optional) The relative percentage of the total number of launched tasks that should use the specified capacity provider. The `weight` value is taken into consideration after the `base` count of tasks has been satisfied. Defaults to `0`.
* `base` - (Optional) The number of tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined. Defaults to `0`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Same as `cluster_name`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS cluster capacity providers using the `cluster_name` attribute. For example:

```terraform
import {
  to = aws_ecs_cluster_capacity_providers.example
  id = "my-cluster"
}
```

Using `terraform import`, import ECS cluster capacity providers using the `cluster_name` attribute. For example:

```console
% terraform import aws_ecs_cluster_capacity_providers.example my-cluster
```
