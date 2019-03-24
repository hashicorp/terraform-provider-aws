---
layout: "aws"
page_title: "AWS: aws_backup_selection"
sidebar_current: "docs-aws-resource-backup-selection"
description: |-
  Provides an AWS Backup selection resource.
---

# aws_backup_selection

Provides an AWS Backup selection to identify resources that are backed up by an AWS Backup plan.

## Example Usage

```hcl
resource "aws_backup_selection" "example" {
  name           = "tf_example_backup_selection"
  backup_plan_id = "${aws_backup_plan.example.id}"
  iam_role_arn   = "${aws_iam_role.example.arn}"

  resources = [
    "${aws_ebs_volume.example.arn}",
    "${aws_efs_file_system.example.arn}",
  ]

  tag_condition {
    test     = "STRINGEQUALS"
    variable = "ec2:ResourceTag/backup"
    value    = "true"
  }

  tag_condition {
    test     = "STRINGEQUALS"
    variable = "ec2:ResourceTag/environment"
    value    = "production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The display name of the backup selection.
* `backup_plan_id` - (Required) Id of the backup plan to associate the backup resources with.
* `iam_role_arn` - (Required) ARN of the role to use for backups and restores.
* `resources` - (Optional) List of resource ARNs to include in the backup selection.
* `tag_condition` - (Optional) A tag condition to include in the backup selection.  Multiple tag_condition blocks may be specified.

### Tag Condition Arguments
For **rule** the following attributes are supported:

* `test` - (Optional) The operator that's applied to this condition.  Currently the only valid value is "STRINGEQUALS" which is also the default.
* `variable` - (Required) - The variable to match against.  Must be a tag key.
* `value` - (Required) - The value to match against.
