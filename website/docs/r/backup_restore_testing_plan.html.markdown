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
  name = "example_restore_testing_plan"

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of the restore testing plan. Must be between 1 and 50 characters long and contain only alphanumeric characters and underscores.
* `recovery_point_selection` - (Required) Recovery point selection configuration. See [`recovery_point_selection`](#recovery_point_selection-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `schedule_expression` - (Required) Schedule expression for the restore testing plan.
* `schedule_expression_timezone` - (Optional) Timezone for the schedule expression. If not provided, the state value will be used.
* `start_window_hours` - (Optional) Number of hours in the start window for the restore testing plan. Must be between 1 and 168.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `recovery_point_selection` Block

* `algorithm` - (Required) Algorithm used for selecting recovery points. Valid values are `RANDOM_WITHIN_WINDOW` and `LATEST_WITHIN_WINDOW`.
* `exclude_vaults` - (Optional) Backup vaults to exclude from the recovery point selection. Each value must be a valid AWS ARN for a backup vault or `*` to exclude all backup vaults.
* `include_vaults` - (Required) Backup vaults to include in the recovery point selection. Each value must be a valid AWS ARN for a backup vault or `*` to include all backup vaults.
* `recovery_point_types` - (Required) Types of recovery points to include in the selection. Valid values are `CONTINUOUS` and `SNAPSHOT`.
* `selection_window_days` - (Optional) Number of days within which the recovery points should be selected. Must be a value between 1 and 365.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Restore Testing Plan.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
