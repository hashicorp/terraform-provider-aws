---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_dashboard"
description: |-
  Provides a CloudWatch Dashboard resource.
---

# Resource: aws_cloudwatch_dashboard

Provides a CloudWatch Dashboard resource.

## Example Usage

```terraform
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "my-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            [
              "AWS/EC2",
              "CPUUtilization",
              "InstanceId",
              "i-012345"
            ]
          ]
          period = 300
          stat   = "Average"
          region = "us-east-1"
          title  = "EC2 Instance CPU"
        }
      },
      {
        type   = "text"
        x      = 0
        y      = 7
        width  = 3
        height = 3

        properties = {
          markdown = "Hello world"
        }
      }
    ]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `dashboard_name` - (Required) The name of the dashboard.
* `dashboard_body` - (Required) The detailed information about the dashboard, including what widgets are included and their location on the dashboard. You can read more about the body structure in the [documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/CloudWatch-Dashboard-Body-Structure.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `dashboard_arn` - The Amazon Resource Name (ARN) of the dashboard.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch dashboards using the `dashboard_name`. For example:

```terraform
import {
  to = aws_cloudwatch_dashboard.sample
  id = "dashboard_name"
}
```

Using `terraform import`, import CloudWatch dashboards using the `dashboard_name`. For example:

```console
% terraform import aws_cloudwatch_dashboard.sample dashboard_name
```
