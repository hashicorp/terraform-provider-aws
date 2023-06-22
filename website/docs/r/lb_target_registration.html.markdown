---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_target_registration"
description: |-
  Terraform resource for managing AWS ELB (Elastic Load Balancing) Target Registration.
---
# Resource: aws_lb_target_registration

Terraform resource for managing AWS ELB (Elastic Load Balancing) Target Registration.

## Example Usage

### Basic Usage

```terraform
resource "aws_lb_target_registration" "example" {
  target_group_arn = aws_lb_target_group.example.arn

  target {
    target_id         = aws_instance.example.id
    port              = 80
    availability_zone = "us-east-1a"
  }
}
```

## Argument Reference

The following arguments are required:

* `target_group_arn` - (Required) Amazon Resource Name (ARN) of the target group.
* `target` - (Required) A list of targets to register with the specified target group. See [`target`](#target).

### `target`

* `target_id` - (Required) The ID of the target. If the target type of the target group is `instance`, specify an instance ID. If the target type is `ip`, specify an IP address. If the target type is `lambda`, specify the ARN of the Lambda function. If the target type is `alb`, specify the ARN of the Application Load Balancer target.
* `availability_zone` - (Optional) An Availability Zone or `all`. This determines whether the target receives traffic from the load balancer nodes in the specified Availability Zone or from all enabled Availability Zones for the load balancer.
* `port` - (Optional) The port on which the target is listening.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier for the target registration.

## Import

AWS ELB (Elastic Load Balancing) Target Registration does not currently support import.
