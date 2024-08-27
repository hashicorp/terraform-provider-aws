---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_target_group_attachment"
description: |-
  Provides the ability to register instances and containers with a LB
  target group
---

# Resource: aws_lb_target_group_attachment

Provides the ability to register instances and containers with an Application Load Balancer (ALB) or Network Load Balancer (NLB) target group. For attaching resources with Elastic Load Balancer (ELB), see the [`aws_elb_attachment` resource](/docs/providers/aws/r/elb_attachment.html).

~> **Note:** `aws_alb_target_group_attachment` is known as `aws_lb_target_group_attachment`. The functionality is identical.

## Example Usage

### Basic Usage

```terraform
resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
  port             = 80
}

resource "aws_lb_target_group" "test" {
  # ... other configuration ...
}

resource "aws_instance" "test" {
  # ... other configuration ...
}
```

### Lambda Target

```terraform
resource "aws_lambda_permission" "with_lb" {
  statement_id  = "AllowExecutionFromlb"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn    = aws_lb_target_group.test.arn
}

resource "aws_lb_target_group" "test" {
  name        = "test"
  target_type = "lambda"
}

resource "aws_lambda_function" "test" {
  # ... other configuration ...
}

resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_lambda_function.test.arn
  depends_on       = [aws_lambda_permission.with_lb]
}
```

### Registering Multiple Targets

```terraform
resource "aws_instance" "example" {
  count = 3
  # ... other configuration ...
}

resource "aws_lb_target_group" "example" {
  # ... other configuration ...
}

resource "aws_lb_target_group_attachment" "example" {
  # covert a list of instance objects to a map with instance ID as the key, and an instance
  # object as the value.
  for_each = {
    for k, v in aws_instance.example :
    k => v
  }

  target_group_arn = aws_lb_target_group.example.arn
  target_id        = each.value.id
  port             = 80
}
```

## Argument Reference

The following arguments are required:

* `target_group_arn` - (Required) The ARN of the target group with which to register targets.
* `target_id` (Required) The ID of the target. This is the Instance ID for an instance, or the container ID for an ECS container. If the target type is `ip`, specify an IP address. If the target type is `lambda`, specify the Lambda function ARN. If the target type is `alb`, specify the ALB ARN.

The following arguments are optional:

* `availability_zone` - (Optional) The Availability Zone where the IP address of the target is to be registered. If the private IP address is outside of the VPC scope, this value must be set to `all`.
* `port` - (Optional) The port on which targets receive traffic.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A unique identifier for the attachment.

## Import

You cannot import Target Group Attachments.
