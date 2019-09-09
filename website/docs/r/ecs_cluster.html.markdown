---
layout: "aws"
page_title: "AWS: aws_ecs_cluster"
sidebar_current: "docs-aws-resource-ecs-cluster"
description: |-
  Provides an ECS cluster.
---

# Resource: aws_ecs_cluster

Provides an ECS cluster.

## Example Usage

```hcl
resource "aws_ecs_cluster" "foo" {
  name = "white-hart"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the cluster (up to 255 letters, numbers, hyphens, and underscores)
* `tags` - (Optional) Key-value mapping of resource tags
* `setting` - (Optional) Configuration block(s) with cluster settings. For example, this can be used to enable CloudWatch Container Insights for a cluster. Defined below.
 
## setting

The `setting` configuration block supports the following:

* `name` - (Required) Name of the setting to manage. Valid values: `containerInsights`.
* `value` -  (Required) The value to assign to the setting. Value values are `enabled` and `disabled`.
 
## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the cluster
* `arn` - The Amazon Resource Name (ARN) that identifies the cluster

## Import

ECS clusters can be imported using the `name`, e.g.

```
$ terraform import aws_ecs_cluster.stateless stateless-app
```
