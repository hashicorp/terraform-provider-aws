---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_selection"
description: |-
  Manages selection conditions for AWS Backup plan resources.
---

# Resource: aws_backup_selection

Manages selection conditions for AWS Backup plan resources.

## Example Usage

### IAM Role

-> For more information about creating and managing IAM Roles for backups and restores, see the [AWS Backup Developer Guide](https://docs.aws.amazon.com/aws-backup/latest/devguide/iam-service-roles.html).

The below example creates an IAM role with the default managed IAM Policy for allowing AWS Backup to create backups.

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["backup.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}
resource "aws_iam_role" "example" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "example" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup"
  role       = aws_iam_role.example.name
}

resource "aws_backup_selection" "example" {
  # ... other configuration ...

  iam_role_arn = aws_iam_role.example.arn
}
```

### Selecting Backups By Tag

```terraform
resource "aws_backup_selection" "example" {
  iam_role_arn = aws_iam_role.example.arn
  name         = "tf_example_backup_selection"
  plan_id      = aws_backup_plan.example.id

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }
}
```

### Selecting Backups By Conditions

```terraform
resource "aws_backup_selection" "example" {
  iam_role_arn = aws_iam_role.example.arn
  name         = "tf_example_backup_selection"
  plan_id      = aws_backup_plan.example.id
  resources    = ["*"]

  condition {
    string_equals {
      key   = "aws:ResourceTag/Component"
      value = "rds"
    }
    string_like {
      key   = "aws:ResourceTag/Application"
      value = "app*"
    }
    string_not_equals {
      key   = "aws:ResourceTag/Backup"
      value = "false"
    }
    string_not_like {
      key   = "aws:ResourceTag/Environment"
      value = "test*"
    }
  }
}
```

### Selecting Backups By Resource

```terraform
resource "aws_backup_selection" "example" {
  iam_role_arn = aws_iam_role.example.arn
  name         = "tf_example_backup_selection"
  plan_id      = aws_backup_plan.example.id

  resources = [
    aws_db_instance.example.arn,
    aws_ebs_volume.example.arn,
    aws_efs_file_system.example.arn,
  ]
}
```

### Selecting Backups By Not Resource

```terraform
resource "aws_backup_selection" "example" {
  iam_role_arn = aws_iam_role.example.arn
  name         = "tf_example_backup_selection"
  plan_id      = aws_backup_plan.example.id

  not_resources = [
    aws_db_instance.example.arn,
    aws_ebs_volume.example.arn,
    aws_efs_file_system.example.arn,
  ]
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The display name of a resource selection document.
* `plan_id` - (Required) The backup plan ID to be associated with the selection of resources.
* `iam_role_arn` - (Required) The ARN of the IAM role that AWS Backup uses to authenticate when restoring and backing up the target resource. See the [AWS Backup Developer Guide](https://docs.aws.amazon.com/aws-backup/latest/devguide/access-control.html#managed-policies) for additional information about using AWS managed policies or creating custom policies attached to the IAM role.
* `selection_tag` - (Optional) Tag-based conditions used to specify a set of resources to assign to a backup plan.
* `condition` - (Optional) A list of conditions that you define to assign resources to your backup plans using tags.
* `resources` - (Optional) An array of strings that either contain Amazon Resource Names (ARNs) or match patterns of resources to assign to a backup plan.
* `not_resources` - (Optional) An array of strings that either contain Amazon Resource Names (ARNs) or match patterns of resources to exclude from a backup plan.

Tag conditions (`selection_tag`) support the following:

* `type` - (Required) An operation, such as `STRINGEQUALS`, that is applied to a key-value pair used to filter resources in a selection.
* `key` - (Required) The key in a key-value pair.
* `value` - (Required) The value in a key-value pair.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Backup Selection identifier

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup selection using the role plan_id and id separated by `|`. For example:

```terraform
import {
  to = aws_backup_selection.example
  id = "plan-id|selection-id"
}
```

Using `terraform import`, import Backup selection using the role plan_id and id separated by `|`. For example:

```console
% terraform import aws_backup_selection.example plan-id|selection-id
```
