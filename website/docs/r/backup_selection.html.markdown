---
layout: "aws"
page_title: "AWS: aws_backup_selection"
sidebar_current: "docs-aws-resource-backup-selection"
description: |-
  Manages selection conditions for AWS Backup plan resources.
---

# aws_backup_selection

Manages selection conditions for AWS Backup plan resources.

## Example Usage

```hcl
resource "aws_backup_selection" "example" {
  plan_id  = "${aws_backup_plan.example.id}"

  name     = "tf_example_backup_selection"
  iam_role = "arn:aws:iam::123456789012:role/service-role/AWSBackupDefaultServiceRole"

  tag {
    type = "STRINGEQUALS"
    key = "foo"
    value = "bar"
  }

  resources = [
    "arn:aws:ec2:us-east-1:123456789012:volume/"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The display name of a resource selection document.
* `plan_id` - (Required) The backup plan ID to be associated with the selection of resources.
* `iam_role` - (Required) The ARN of the IAM role that AWS Backup uses to authenticate when restoring the target resource.
* `selection_tag` - (Optional) Tag-based conditions used to specify a set of resources to assign to a backup plan.
* `resources` - (Optional) An array of strings that either contain Amazon Resource Names (ARNs) or match patterns of resources to assign to a backup plan..

Tag conditions (`selection_tag`) support the following:

* `type` - (Required) An operation, such as `StringEquals`, that is applied to a key-value pair used to filter resources in a selection.
* `key` - (Required) The key in a key-value pair.
* `value` - (Required) The value in a key-value pair.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Backup Selection identifier
