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

### Target using QUIC

```terraform
resource "aws_lb_target_group" "test" {
  name     = "test"
  port     = 443
  protocol = "QUIC"
  # ... other configuration ...
}

resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
  port             = 443
  quic_server_id   = "0x1a2b3c4d5e6f7a8b"
}

resource "aws_instance" "test" {
  # ... other configuration ...
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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `availability_zone` - (Optional) The Availability Zone where the IP address of the target is to be registered. If the private IP address is outside of the VPC scope, this value must be set to `all`.
* `port` - (Optional) The port on which targets receive traffic.
* `quic_server_id` - (Optional) Server ID for the targets, consisting of the 0x prefix followed by 16 hexadecimal characters. The value must be unique at the listener level. Required if `aws_lb_target_group` protocol is `QUIC` or `TCP_QUIC`. Not valid with other protocols. Forces replacement if modified.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A unique identifier for the attachment.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lb_target_group_attachment.example
  identity = {
    target_group_arn = "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-tg/abc123"
    target_id        = "i-0123456789abcdef0"
    port             = 8080
  }
}

resource "aws_lb_target_group_attachment" "example" {
  target_group_arn = "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-tg/abc123"
  target_id        = "i-0123456789abcdef0"
  port             = 8080
}
```

### Identity Schema

#### Required

* `target_group_arn` - (String) ARN of the target group.
* `target_id` - (String) ID of the target (instance ID, IP address, Lambda ARN, or ALB ARN).

#### Optional

* `port` - (Number) Port on which targets receive traffic.
* `availability_zone` - (String) Availability zone where the target is registered.
* `account_id` - (String) AWS Account where this resource is managed.
* `region` - (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Target Group Attachments using `target_group_arn`, `target_id`, and optionally `port` and `availability_zone` separated by commas (`,`). For example:

```terraform
import {
  to = aws_lb_target_group_attachment.example
  id = "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-tg/abc123,i-0123456789abcdef0,8080"
}
```

Using `terraform import`, import Target Group Attachments using the same format. For example:

```console
% terraform import aws_lb_target_group_attachment.example arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-tg/abc123,i-0123456789abcdef0,8080
```
