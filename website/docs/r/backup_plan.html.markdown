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

* `name` - (Required) The display name of a backup plan.
* `rule` - (Required) A rule object that specifies a scheduled task that is used to back up a selection of resources.
* `advanced_backup_setting` - (Optional) An object that specifies backup options for each resource type.
* `tags` - (Optional) Metadata that you can assign to help organize the plans you create. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Rule Arguments

`rule` supports the following attributes:

* `rule_name` - (Required) An display name for a backup rule.
* `target_vault_name` - (Required) The name of a logical container where backups are stored.
* `schedule` - (Optional) A CRON expression specifying when AWS Backup initiates a backup job.
* `enable_continuous_backup` - (Optional) Enable continuous backups for supported resources.
* `start_window` - (Optional) The amount of time in minutes before beginning a backup.
* `completion_window` - (Optional) The amount of time in minutes AWS Backup attempts a backup before canceling the job and returning an error.
* `lifecycle` - (Optional) The lifecycle defines when a protected resource is transitioned to cold storage and when it expires.  Fields documented below.
* `recovery_point_tags` - (Optional) Metadata that you can assign to help organize the resources that you create.
* `copy_action` - (Optional) Configuration block(s) with copy operation settings. Detailed below.

### Lifecycle Arguments

`lifecycle` supports the following attributes:

* `cold_storage_after` - (Optional) Specifies the number of days after creation that a recovery point is moved to cold storage.
* `delete_after` - (Optional) Specifies the number of days after creation that a recovery point is deleted. Must be 90 days greater than `cold_storage_after`.
* `opt_in_to_archive_for_supported_resources` - (Optional) This setting will instruct your backup plan to transition supported resources to archive (cold) storage tier in accordance with your lifecycle settings.

### Copy Action Arguments

`copy_action` supports the following attributes:

* `lifecycle` - (Optional) The lifecycle defines when a protected resource is copied over to a backup vault and when it expires.  Fields documented above.
* `destination_vault_arn` - (Required) An Amazon Resource Name (ARN) that uniquely identifies the destination backup vault for the copied backup.

### Advanced Backup Setting Arguments

`advanced_backup_setting` supports the following arguments:

* `backup_options` - (Required) Specifies the backup option for a selected resource. This option is only available for Windows VSS backup jobs. Set to `{ WindowsVSS = "enabled" }` to enable Windows VSS backup option and create a VSS Windows backup.
* `resource_type` - (Required) The type of AWS resource to be backed up. For VSS Windows backups, the only supported resource type is Amazon EC2. Valid values: `EC2`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The id of the backup plan.
* `arn` - The ARN of the backup plan.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - Unique, randomly generated, Unicode, UTF-8 encoded string that serves as the version ID of the backup plan.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Plan using the `id`. For example:

```terraform
import {
  to = aws_backup_plan.test
  id = "<id>"
}
```

Using `terraform import`, import Backup Plan using the `id`. For example:

```console
% terraform import aws_backup_plan.test <id>
```
