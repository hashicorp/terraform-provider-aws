---
subcategory: "CloudWatch RUM"
layout: "aws"
page_title: "AWS: aws_rum_resource_policy"
description: |-
  Provides a CloudWatch RUM Resource Policy resource.
---

# Resource: aws_rum_resource_policy

Provides a CloudWatch RUM Resource Policy resource.

## Example Usage

```terraform
resource "aws_rum_app_monitor" "example" {
  name   = "example"
  domain = "localhost"
}

resource "aws_rum_resource_policy" "example" {
  app_monitor_name = aws_rum_app_monitor.example.name
  policy_document  = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowRUM",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "rum:PutRumMetricsDestination",
      "Resource": "*"
    }
  ]
}
POLICY
}
```

## Argument Reference

This resource supports the following arguments:

* `app_monitor_name` - (Required) The name of the CloudWatch RUM App Monitor.
* `policy_document` - (Required) The JSON policy document.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the CloudWatch RUM App Monitor.
* `policy_revision_id` - The revision ID of the policy.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cloudwatch RUM Resource Policy using the `app_monitor_name`. For example:

```terraform
import {
  to = aws_rum_resource_policy.example
  id = "example"
}
```

Using `terraform import`, import Cloudwatch RUM Resource Policy using the `app_monitor_name`. For example:

```console
% terraform import aws_rum_resource_policy.example example
```
