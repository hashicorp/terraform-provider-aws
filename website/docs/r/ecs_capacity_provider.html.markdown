---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_capacity_provider"
description: |-
  Provides an ECS capacity provider.
---

# Resource: aws_ecs_capacity_provider

Provides an ECS capacity provider - a way of enabling autoscaling for ECS clusters. 

## Example Usage

```hcl
resource "aws_ecs_capacity_provider" "test" {
  name = "test" 
  tags = {}

  auto_scaling_group_provider {        
    auto_scaling_group_arn = "${aws_autoscaling_group.test.arn}"  
    managed_termination_protection = "ENABLED" 

    managed_scaling {
      maximum_scaling_step_size = 1000
      minimum_scaling_step_size = 1
      status = "ENABLED"
      target_capacity = 10
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the capacity provider
* `auto_scaling_group_provider` - (Required) Nested argument defining the provider for the ECS auto-scaling group. Defined below.
* `tags` - (Optional) Key-value mapping of resource tags

## auto_scaling_group_provider

The `auto_scaling_group_provider` supports the following:

* `auto_scaling_group_arn` - (Required) - The ARN of the associated auto-scaling group
* `managed_termination_protection` - (Required) - Enables or disables container-aware termination of instances in the ASG when scale-in happens
* `managed_scaling` - (Required) - Nested argument defining the parameters of the auto scaling. Defined below.

## managed_scaling

The `managed_scaling` block supports the following:

* `status` - (Required) Whether or not the autoscaling is currently `ENABLED` or `DISABLED`
* `target_capacity` - (Required) The target number of ECS instances in the ASG
* `maximum_scaling_step_size` - (Required) The maximum step adjustment size
* `minimum_scaling_step_size` - (Required) The minimum step adjustment size

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the service