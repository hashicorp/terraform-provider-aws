---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_daemons"
description: |-
    Provides a list of ECS daemons in a cluster
---

# Data Source: aws_ecs_daemons

The ECS daemons data source allows access to a list of AWS ECS daemons in a specific cluster.

## Example Usage

```terraform
data "aws_ecs_daemons" "all" {
  cluster = "arn:aws:ecs:us-west-2:123456789012:cluster/my-cluster"
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster` - (Required) ARN of the ECS cluster to list daemons for.
* `capacity_provider_arns` - (Optional) List of capacity provider ARNs to filter the daemons by.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `daemon_arns` - List of daemon ARNs in the cluster.
* `id` - ARN of the ECS cluster.
