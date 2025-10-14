---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_maintenance_windows"
description: |-
  Get information on SSM maintenance windows.
---

# Data Source: aws_ssm_maintenance_windows

Use this data source to get the window IDs of SSM maintenance windows.

## Example Usage

```terraform
data "aws_ssm_maintenance_windows" "example" {
  filter {
    name   = "Enabled"
    values = ["true"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [SSM DescribeMaintenanceWindows API Reference](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_DescribeMaintenanceWindows.html#API_DescribeMaintenanceWindows_RequestSyntax).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of window IDs of the matched SSM maintenance windows.
