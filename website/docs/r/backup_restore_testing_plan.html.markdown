---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_restore_testing_plan"
description: |-
  Provides an AWS Backup Restore Testing plan resource.
---

# Resource: aws_backup_restore_testing_plan

Provides an AWS Backup Restore Testing plan resource.

## Example Usage

```terraform
resource "aws_backup_restore_testing_plan" "example" {
  name = "tf_example_restore_testing_plan"
  schedule          = "cron(0 12 * * ? *)"

  recovery_point_selection {
    algorithm = "RANDOM_WITHIN_WINDOW"
    include_vaults = ["*"]
    recovery_point_types = ["CONTINUOUS", "SNAPSHOT"]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The display name of a restore testing plan.\
* `recovery_point_selection` - (Required) A recovery point selection object that specifies the selection of protected resources and the method when perform restores.
* `schedule` - (Required) A CRON expression specifying when AWS Backup initiates a restore testing job.
* `schedule_timezone` - (Optional) The timezone of the schedule. Defaults to `UTC`
* `start_window` - (Optional) The number of hours before cancelling a job if it doesn't start successfully. Defaults to `24`
* `tags` - (Optional) Metadata that you can assign to help organize the plans you create. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Recovery Point Selection Arguments

`recovery_point_selection` supports the following attributes:

* `algorithm` - (Required) The algorithm used when restoring a recovery point. For more information, see [RestoreTestingRecoveryPointSelection](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_RecoveryPointSelector.html).
* `include_vaults` - (Required) An array of strings that contains ARNs of vaults which are included for selection. For more information, see [RestoreTestingRecoveryPointSelection](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_RecoveryPointSelector.html).
* `recovery_point_types` - (Required) An array of strings that contains the types of recovery points to select. For more information, see [RestoreTestingRecoveryPointSelection](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_RecoveryPointSelector.html).
* `exclude_vaults` - (Optional) An array of strings that contains ARNs of vaults which are excluded for selection.
* `selection_window` - (Optional) The window in days for the selection of eligible recovery points. Defaults to `30`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The id of the restore testing plan.
* `arn` - The ARN of the restore testing plan.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Restore Testing Plan using the `id`. For example:

```terraform
import {
  to = aws_backup_restore_testing_plan.example
  id = "<id>"
}
```

Using `terraform import`, import Restore Testing Plan using the `id`. For example:

```console
% terraform import aws_backup_restore_testing_plan.example <id>
```
