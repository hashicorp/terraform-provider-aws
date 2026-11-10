---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_connection_notification"
description: |-
  Provides a VPC Endpoint connection notification resource.
---

# Resource: aws_vpc_endpoint_connection_notification

Provides a VPC Endpoint connection notification resource.
Connection notifications notify subscribers of VPC Endpoint events.

## Example Usage

```terraform
data "aws_iam_policy_document" "topic" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["vpce.amazonaws.com"]
    }

    actions   = ["SNS:Publish"]
    resources = ["arn:aws:sns:*:*:vpce-notification-topic"]
  }
}

resource "aws_sns_topic" "topic" {
  name   = "vpce-notification-topic"
  policy = data.aws_iam_policy_document.topic.json
}

resource "aws_vpc_endpoint_service" "foo" {
  acceptance_required        = false
  network_load_balancer_arns = [aws_lb.test.arn]
}

resource "aws_vpc_endpoint_connection_notification" "foo" {
  vpc_endpoint_service_id     = aws_vpc_endpoint_service.foo.id
  connection_notification_arn = aws_sns_topic.topic.arn
  connection_events           = ["Accept", "Reject"]
}
```

## Argument Reference

This resource supports the following arguments:

* `connection_events` - (Required) One or more endpoint [events](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateVpcEndpointConnectionNotification.html#API_CreateVpcEndpointConnectionNotification_RequestParameters) for which to receive notifications.
* `connection_notification_arn` - (Required) ARN of the SNS topic for the notifications.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_endpoint_id` - (Optional) ID of the VPC Endpoint to receive notifications for.
* `vpc_endpoint_service_id` - (Optional) ID of the VPC Endpoint Service to receive notifications for.

~> **NOTE:** One of `vpc_endpoint_service_id` or `vpc_endpoint_id` must be specified.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the VPC connection notification.
* `notification_type` - Type of notification.
* `state` - State of the notification.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Endpoint connection notifications using the VPC endpoint connection notification `id`. For example:

```terraform
import {
  to = aws_vpc_endpoint_connection_notification.foo
  id = "vpce-nfn-09e6ed3b4efba2263"
}
```

Using `terraform import`, import VPC Endpoint connection notifications using the VPC endpoint connection notification `id`. For example:

```console
% terraform import aws_vpc_endpoint_connection_notification.foo vpce-nfn-09e6ed3b4efba2263
```
