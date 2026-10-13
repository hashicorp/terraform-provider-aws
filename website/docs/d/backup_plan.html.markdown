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
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the backup plan.
* `name` - Display name of a backup plan.
* `rule` - Rules of a backup plan. [See below](#rule-block).
* `scan_setting` - Scanning configuration for the backup rule. [See below](#scan_setting-block).
* `tags` - Metadata that you can assign to help organize the plans you create.
* `version` - Unique, randomly generated, Unicode, UTF-8 encoded string that serves as the version ID of the backup plan.

### `rule` Block

* `completion_window` - Amount of time in minutes AWS Backup attempts a backup before canceling the job and returning an error.
* `copy_action` - Configuration block(s) with copy operation settings. [See below](#copy_action-block).
* `enable_continuous_backup` - Whether AWS Backup creates continuous backups.
* `lifecycle` - Lifecycle defining when a recovery point transitions to cold storage and when it expires. [See below](#lifecycle-block).
* `recovery_point_tags` - Metadata that you can assign to help organize the resources that you create.
* `rule_name` - Display name of a backup rule.
* `scan_action` - Configuration block(s) with malware scanning settings. [See below](#scan_action-block).
* `schedule` - CRON expression specifying when AWS Backup initiates a backup job.
* `schedule_expression_timezone` - Timezone in which the schedule expression is set.
* `start_window` - Amount of time in minutes before beginning a backup.
* `target_logically_air_gapped_backup_vault_arn` - ARN of the logically air-gapped backup vault where the recovery point is copied.
* `target_vault_name` - Name of a logical container where backups are stored.

#### `copy_action` Block

* `destination_vault_arn` - ARN of the destination backup vault for the copied backup.
* `lifecycle` - Lifecycle defining when a recovery point transitions to cold storage and when it expires. [See below](#lifecycle-block).

#### `lifecycle` Block

* `cold_storage_after` - Number of days after creation that a recovery point is moved to cold storage.
* `delete_after` - Number of days after creation that a recovery point is deleted.
* `opt_in_to_archive_for_supported_resources` - Whether the recovery point is transitioned to cold storage for supported resource types.

#### `scan_action` Block

* `malware_scanner` - Malware scanner used for the scan action.
* `scan_mode` - Mode of the malware scan.

### `scan_setting` Block

* `malware_scanner` - Malware scanner used for the scan setting.
* `resource_types` - Resource types to scan.
* `scanner_role_arn` - ARN of the IAM role used by the scanner.
