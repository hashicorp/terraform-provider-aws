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

### Selecting Backups By Resource Type And Tag

~> **Note:** To select resources that match both an ARN pattern (for example a service-scoped wildcard) **and** a tag, use `condition` with `resources`. Do **not** combine `resources` wildcards with `selection_tag` — that combination uses OR semantics and can select far more resources than expected. See [Argument Reference](#argument-reference) below.

```terraform
resource "aws_backup_selection" "example" {
  iam_role_arn = aws_iam_role.example.arn
  name         = "tf_example_backup_selection"
  plan_id      = aws_backup_plan.example.id

  resources = ["arn:aws:s3:::*"]

  condition {
    string_equals {
      key   = "aws:ResourceTag/enable_backup"
      value = "true"
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

~> **Note:** `resources` and `selection_tag` are evaluated independently and combined with **OR** semantics (AWS Backup `Resources` + `ListOfTags`). A resource is selected if it matches **any** ARN/pattern in `resources` **or** **any** `selection_tag`. Tag conditions in `selection_tag` do **not** further filter the ARNs in `resources`. In particular, `resources = ["*"]` or a service-scoped wildcard such as `["arn:aws:s3:::*"]` together with `selection_tag` selects all resources matching the ARN pattern **plus** all resources matching the tag(s), which can result in an unintended account-wide selection. To require **both** an ARN match and a tag match, set `resources` to the desired pattern and use `condition` (AWS Backup `Conditions`) instead of `selection_tag`. Multiple `condition` rules use **AND** semantics. See [CreateBackupSelection](https://docs.aws.amazon.com/cli/latest/reference/backup/create-backup-selection.html).

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The display name of a resource selection document.
* `plan_id` - (Required) The backup plan ID to be associated with the selection of resources.
* `iam_role_arn` - (Required) The ARN of the IAM role that AWS Backup uses to authenticate when restoring and backing up the target resource. See the [AWS Backup Developer Guide](https://docs.aws.amazon.com/aws-backup/latest/devguide/access-control.html#managed-policies) for additional information about using AWS managed policies or creating custom policies attached to the IAM role.
* `selection_tag` - (Optional) Tag-based conditions used to specify a set of resources to assign to a backup plan (`ListOfTags` in the AWS API). Multiple tags use OR logic among themselves. When set together with `resources`, selection is the union of both (OR), not an intersection. See [below](#selection_tag) for details.
* `condition` - (Optional) Condition-based filters applied to resources selected for a backup plan (`Conditions` in the AWS API). Multiple conditions use AND logic. Use this with `resources` when you need tag filters to narrow an ARN pattern. See [below](#condition) for details.
* `resources` - (Optional) Array of strings that either contain ARNs or match patterns of resources to assign to a backup plan. Multiple ARNs use OR logic. Combined with `selection_tag` using OR semantics (see note above).
* `not_resources` - (Optional) Array of strings that either contain ARNs or match patterns of resources to exclude from a backup plan.

### selection_tag

The `selection_tag` configuration block supports the following attributes:

* `type` - (Required) An operation, such as `STRINGEQUALS`, that is applied to the key-value pair used to filter resources in a selection.
* `key` - (Required) Key for the filter.
* `value` - (Required) Value for the filter.

### condition

The `condition` configuration block supports the following attributes:

* `string_equals` - (Optional) Filters the values of your tagged resources for only those resources that you tagged with the same value. Also called "exact matching". See [below](#string_equals) for details.
* `string_not_equals` - (Optional) Filters the values of your tagged resources for only those resources that you tagged that do not have the same value. Also called "negated matching". See [below](#string_not_equals) for details.
* `string_like` - (Optional) Filters the values of your tagged resources for matching tag values with the use of a wildcard character (`*`) anywhere in the string. For example, `prod*` or `*rod*` matches the tag value `production`. See [below](#string_like) for details.
* `string_not_like` - (Optional) Filters the values of your tagged resources for non-matching tag values with the use of a wildcard character (`*`) anywhere in the string. See [below](#string_not_like) for details.

### string_equals

The `string_equals` configuration block supports the following attributes:

* `key` - (Required) Key for the filter.
* `value` - (Required) Value for the filter.

### string_not_equals

The `string_not_equals` configuration block supports the following attributes:

* `key` - (Required) Key for the filter.
* `value` - (Required) Value for the filter.

### string_like

The `string_like` configuration block supports the following attributes:

* `key` - (Required) Key for the filter.
* `value` - (Required) Value for the filter.

### string_not_like

The `string_not_like` configuration block supports the following attributes:

* `key` - (Required) Key for the filter.
* `value` - (Required)  Value for the filter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Backup Selection identifier.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_backup_selection.example
  identity = {
    plan_id = "abcd1234"
    id      = "efgh5678"
  }
}

resource "aws_backup_selection" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `plan_id` (Required) Backup plan ID associated with the selection of resources.
* `id` (String) Backup selection ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup selection using the `plan_id` and `id` separated by `|`. For example:

```terraform
import {
  to = aws_backup_selection.example
  id = "abcd1234|efgh5678"
}
```

Using `terraform import`, import Backup selection using the `plan_id` and `id` separated by `|`. For example:

```console
% terraform import aws_backup_selection.example abcd1234|efgh5678
```
