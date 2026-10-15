---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_plan"
description: |-
  Provides an AWS Backup plan resource.
---

# Resource: aws_backup_plan

Provides an AWS Backup plan resource.

## Example Usage

```terraform
resource "aws_backup_plan" "example" {
  name = "tf_example_backup_plan"

  rule {
    rule_name         = "tf_example_backup_rule"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"

    lifecycle {
      delete_after = 14
    }
  }

  advanced_backup_setting {
    backup_options = {
      WindowsVSS = "enabled"
    }
    resource_type = "EC2"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `advanced_backup_setting` - (Optional) Object that specifies backup options for each resource type. Detailed below.
* `name` - (Required) Display name of a backup plan.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rule` - (Required) Rule that specifies a scheduled task used to back up a selection of resources. Detailed below.
* `scan_setting` - (Optional) Block for scanning configuration for the backup rule and includes the malware scanner, and scan mode of either full or incremental. Detailed below.
* `tags` - (Optional) Metadata that you can assign to help organize the plans you create. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `rule` Block

`rule` supports the following attributes:

* `completion_window` - (Optional) Amount of time in minutes AWS Backup attempts a backup before canceling the job and returning an error.
* `copy_action` - (Optional) Configuration block(s) with copy operation settings. Detailed below.
* `enable_continuous_backup` - (Optional) Enable continuous backups for supported resources.
* `lifecycle` - (Optional) Lifecycle that defines when a protected resource is transitioned to cold storage and when it expires. Detailed below.
* `recovery_point_tags` - (Optional) Metadata that you can assign to help organize the resources that you create.
* `rule_name` - (Required) Display name for a backup rule.
* `scan_action` - (Optional) Block for scanning configuration for the backup rule and includes the malware scanner, and scan mode of either full or incremental. Detailed below.
* `schedule` - (Optional) CRON expression specifying when AWS Backup initiates a backup job.
* `schedule_expression_timezone` - (Optional) Timezone in which the schedule expression is set. Default value: `"Etc/UTC"`.
* `start_window` - (Optional) Amount of time in minutes before beginning a backup.
* `target_logically_air_gapped_backup_vault_arn` - (Optional) ARN of a logically air-gapped vault. ARN must be in the same account and region. If provided, supported fully managed resources back up directly to logically air-gapped vault, while other supported resources create a temporary (billable) snapshot in backup vault, then copy it to logically air-gapped vault. Unsupported resources only back up to the specified backup vault.
* `target_vault_name` - (Required) Name of a logical container where backups are stored.

### `copy_action` Block

`copy_action` supports the following attributes:

* `destination_vault_arn` - (Required) ARN that uniquely identifies the destination backup vault for the copied backup.
* `lifecycle` - (Optional) Lifecycle that defines when a protected resource is copied over to a backup vault and when it expires. Detailed below.

### `lifecycle` Block

`lifecycle` supports the following attributes:

* `cold_storage_after` - (Optional) Number of days after creation that a recovery point is moved to cold storage.
* `delete_after` - (Optional) Number of days after creation that a recovery point is deleted. Must be 90 days greater than `cold_storage_after`.
* `opt_in_to_archive_for_supported_resources` - (Optional) Whether to transition supported resources to archive (cold) storage tier in accordance with your lifecycle settings.

### `scan_action` Block

`scan_action` supports the following attributes:

* `malware_scanner` - (Required) Malware scanner to use for the scan action. Currently only `GUARDDUTY` is supported.
* `scan_mode` - (Required) Scanning mode to use for the scan action. Valid values are `FULL_SCAN` and `INCREMENTAL_SCAN`.

### `advanced_backup_setting` Block

`advanced_backup_setting` supports the following arguments:

* `backup_options` - (Required) Backup option for a selected resource. This option is only available for Windows VSS backup jobs. Set to `{ WindowsVSS = "enabled" }` to enable Windows VSS backup option and create a VSS Windows backup.
* `resource_type` - (Required) Type of AWS resource to be backed up. For VSS Windows backups, the only supported resource type is Amazon EC2. Valid values: `EC2`.

### `scan_setting` Block

`scan_setting` supports the following attributes:

* `malware_scanner` - (Required) Malware scanner to use for the scan setting. Currently only `GUARDDUTY` is supported.
* `resource_types` - (Required) List of resource types to apply the scan setting to. Valid values are `EBS`, `EC2`, `S3` and `ALL`.
* `scanner_role_arn` - (Required) ARN of the IAM role that AWS Backup uses to scan resources. See [the AWS documentation](https://docs.aws.amazon.com/guardduty/latest/ug/malware-protection-backup-iam-permissions.html) for details.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the backup plan.
* `id` - ID of the backup plan.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - Unique, randomly generated, Unicode, UTF-8 encoded string that serves as the version ID of the backup plan.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_backup_plan.example
  identity = {
    id = "abc123"
  }
}

resource "aws_backup_plan" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) Backup Plan ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Plan using the `id`. For example:

```terraform
import {
  to = aws_backup_plan.example
  id = "abc123"
}
```

Using `terraform import`, import Backup Plan using the `id`. For example:

```console
% terraform import aws_backup_plan.example abc123
```
