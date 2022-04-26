---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_group"
description: |-
  Get information on a Cloudwatch Log Group.
---

# Data Source: aws_cloudwatch_log_group

Use this data source to get information about an AWS Cloudwatch Log Group

## Example Usage

```terraform
data "aws_cloudwatch_log_group" "example" {
  name = "MyImportantLogs"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Cloudwatch log group

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Cloudwatch log group. Any `:*` suffix added by the API, denoting all CloudWatch Log Streams under the CloudWatch Log Group, is removed for greater compatibility with other AWS services that do not accept the suffix.
* `creation_time` - The creation time of the log group, expressed as the number of milliseconds after Jan 1, 1970 00:00:00 UTC.
* `retention_in_days` - The number of days log events retained in the specified log group.
* `kms_key_id` - The ARN of the KMS Key to use when encrypting log data.
* `tags` - A map of tags to assign to the resource.
