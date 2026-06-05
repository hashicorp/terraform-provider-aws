---
subcategory: "BCM Dashboards"
layout: "aws"
page_title: "AWS: aws_bcmdashboards_dashboard"
description: |-
  Terraform data source for an AWS Billing and Cost Management Dashboards Dashboard.
---

# Data Source: aws_bcmdashboards_dashboard

Terraform data source for an AWS Billing and Cost Management (BCM) Dashboards Dashboard.

## Example Usage

### Basic Usage

```terraform
data "aws_bcmdashboards_dashboard" "example" {
  arn = "arn:aws:bcm-dashboards::123456789012:dashboard/01234567-89ab-cdef-0123-456789abcdef"
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) ARN of the dashboard.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `created_at` - Timestamp when the dashboard was created.
* `dashboard_type` - Type of the dashboard.
* `description` - Description of the dashboard.
* `name` - Name of the dashboard.
* `tags` - Map of tags assigned to the dashboard.
* `updated_at` - Timestamp when the dashboard was last updated.
* `widget` - Widgets that make up the dashboard. The structure mirrors the [`widget` argument of the `aws_bcmdashboards_dashboard` resource](aws_bcmdashboards_dashboard.html.markdown#widget).
