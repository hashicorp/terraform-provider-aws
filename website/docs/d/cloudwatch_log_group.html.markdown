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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the Cloudwatch log group

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Cloudwatch log group. Any `:*` suffix added by the API, denoting all CloudWatch Log Streams under the CloudWatch Log Group, is removed for greater compatibility with other AWS services that do not accept the suffix.
* `creation_time` - Creation time of the log group, expressed as the number of milliseconds after Jan 1, 1970 00:00:00 UTC.
* `kms_key_id` - ARN of the KMS Key to use when encrypting log data.
* `log_group_class` - The log class of the log group.
* `retention_in_days` - Number of days log events retained in the specified log group.
* `tags` - Map of tags to assign to the resource.
