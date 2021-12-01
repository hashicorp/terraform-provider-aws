---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_protection_health_check_association"
description: |-
  Creates an association between a Route53 Health Check and a Shield Advanced protected resource.
---

# Resource: aws_shield_protection_health_check_association

Creates an association between a Route53 Health Check and a Shield Advanced protected resource.
This association uses the health of your applications to improve responsiveness and accuracy in attack detection and mitigation.

Blog post: [AWS Shield Advanced now supports Health Based Detection](https://aws.amazon.com/about-aws/whats-new/2020/02/aws-shield-advanced-now-supports-health-based-detection/)

## Example Usage

### Create an association between a protected EIP and a Route53 Health Check

```terraform
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_eip" "example" {
  vpc = true
  tags = {
    Name = "example"
  }
}

resource "aws_shield_protection" "example" {
  name         = "example-protection"
  resource_arn = "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.example.id}"
}

resource "aws_route53_health_check" "example" {
  ip_address        = aws_eip.example.public_ip
  port              = 80
  type              = "HTTP"
  resource_path     = "/ready"
  failure_threshold = "3"
  request_interval  = "30"

  tags = {
    Name = "tf-example-health-check"
  }
}

resource "aws_shield_protection_health_check_association" "example" {
  health_check_arn     = aws_route53_health_check.example.arn
  shield_protection_id = aws_shield_protection.example.id
}
```

## Argument Reference

The following arguments are supported:

* `health_check_arn` - (Required) The ARN (Amazon Resource Name) of the Route53 Health Check resource which will be associated to the protected resource.
* `shield_protection_id` - (Required) The ID of the protected resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) for the Protection object that is created.

## Import

Shield protection health check association resources can be imported by specifying the `shield_protection_id` and `health_check_arn` e.g.,

```
$ terraform import aws_shield_protection_health_check_association.example ff9592dc-22f3-4e88-afa1-7b29fde9669a+arn:aws:route53:::healthcheck/3742b175-edb9-46bc-9359-f53e3b794b1b
```
