---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_plan"
description: |-
  Provides details about an AWS Backup plan.
---

# Data Source: aws_backup_plan

Use this data source to get information on an existing backup plan.

## Example Usage

```terraform
data "aws_backup_plan" "example" {
  plan_id = "tf_example_backup_plan_id"
}
```

## Argument Reference

This data source supports the following arguments:

* `plan_id` - (Required) Backup plan ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the backup plan.
* `name` - Display name of a backup plan.
* `rule` - Rules of a backup plan.
* `tags` - Metadata that you can assign to help organize the plans you create.
* `version` - Unique, randomly generated, Unicode, UTF-8 encoded string that serves as the version ID of the backup plan.
