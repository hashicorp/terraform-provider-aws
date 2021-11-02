---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatchlogs_groups"
description: |-
  Get list of Cloudwatch Log Groups.
---

# Data Source: aws_cloudwatchlogs_groups

Use this data source to get a list of AWS Cloudwatch Log Groups

## Example Usage

```terraform
data "aws_cloudwatchlogs_groups" "example" {
  log_group_name_prefix = "/MyImportantLogs"
}
```

## Argument Reference

The following arguments are supported:

* `log_group_name_prefix` - (Required) The group prefix of the Cloudwatch log groups to list

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arns` - Set of ARNs of the Cloudwatch log groups
* `log_group_names` - Set of names of the Cloudwatch log groups
