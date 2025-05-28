---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_groups"
description: |-
  Get list of Cloudwatch Log Groups.
---

# Data Source: aws_cloudwatch_log_groups

Use this data source to get a list of AWS Cloudwatch Log Groups

## Example Usage

```terraform
data "aws_cloudwatch_log_groups" "example" {
  log_group_name_prefix = "/MyImportantLogs"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `log_group_name_prefix` - (Optional) Group prefix of the Cloudwatch log groups to list

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of ARNs of the Cloudwatch log groups
* `log_group_names` - Set of names of the Cloudwatch log groups
