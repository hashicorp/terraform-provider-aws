---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_restore_testing_plan"
description: |-
  Terraform resource for managing an AWS Backup Restore Testing Plan.
---
# Resource: aws_backup_restore_testing_plan

Terraform resource for managing an AWS Backup Restore Testing Plan.

## Example Usage

### Basic Usage

```terraform
resource "aws_backup_restore_testing_plan" "example" {
  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00
}
```

## Argument Reference

The following arguments are required:

* `name` (Required): The name of the restore testing plan. Must be between 1 and 50 characters long and contain only alphanumeric characters and underscores.
* `schedule_expression` (Required): The schedule expression for the restore testing plan.
* `schedule_expression_timezone` (Optional): The timezone for the schedule expression. If not provided, the state value will be used.
* `start_window_hours` (Optional): The number of hours in the start window for the restore testing plan. Must be between 1 and 168.
* `recovery_point_selection` (Required): Specifies the recovery point selection configuration. See [RecoveryPointSelection](#recoverypointselection) section for more details.

### RecoveryPointSelection

* `algorithm` (Required): Specifies the algorithm used for selecting recovery points. Valid values are "RANDOM_WITHIN_WINDOW" and "LATEST_WITHIN_WINDOW".
* `include_vaults` (Required): Specifies the backup vaults to include in the recovery point selection. Each value must be a valid AWS ARN for a backup vault or "*" to include all backup vaults.
* `recovery_point_types` (Required): Specifies the types of recovery points to include in the selection. Valid values are "CONTINUOUS" and "SNAPSHOT".
* `exclude_vaults` (Optional): Specifies the backup vaults to exclude from the recovery point selection. Each value must be a valid AWS ARN for a backup vault or "*" to exclude all backup vaults.
* `selection_window_days` (Optional): Specifies the number of days within which the recovery points should be selected. Must be a value between 1 and 365.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Restore Testing Plan.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Restore Testing Plan using the `name`. For example:

```terraform
import {
  to = aws_backup_restore_testing_plan.example
  id = "my_testing_plan"
}
```

Using `terraform import`, import Backup Restore Testing Plan using the `name`. For example:

```console
% terraform import aws_backup_restore_testing_plan.example my_testing_plan
```
