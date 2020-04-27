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

```hcl
data "aws_backup_plan" "example" {
  name = "tf_example_backup_plan"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the backup plan.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `rule` - A rule object that specifies a scheduled task that is used to back up a selection of resources.
* `tags` - Metadata that you can assign to help organize the plans you create.
* `arn` - The ARN of the backup plan.
* `version` - Unique, randomly generated, Unicode, UTF-8 encoded string that serves as the version ID of the backup plan.

### Rule Arguments
For **rule** the following attributes are available:

* `rule_name` - A display name for a backup rule.
* `target_vault_name` - The name of a logical container where backups are stored.
* `schedule` - A CRON expression specifying when AWS Backup initiates a backup job.
* `start_window` - The amount of time in minutes before beginning a backup.
* `completion_window` - The amount of time AWS Backup attempts a backup before canceling the job and returning an error.
* `lifecycle` - The lifecycle defines when a protected resource is transitioned to cold storage and when it expires.  Fields documented below.
* `recovery_point_tags` - Metadata that you can assign to help organize the resources that you create.
* `copy_action` - Configuration block(s) with copy operation settings. Detailed below.

### Lifecycle Arguments
For **lifecycle** the following attributes are available:

* `cold_storage_after` - Specifies the number of days after creation that a recovery point is moved to cold storage.
* `delete_after` - Specifies the number of days after creation that a recovery point is deleted. Must be 90 days greater than `cold_storage_after`.

### Copy Action Arguments
For **copy_action** the following attributes are available:

* `lifecycle` - The lifecycle defines when a protected resource is copied over to a backup vault and when it expires.  Fields documented above.
* `destination_vault_arn` - An Amazon Resource Name (ARN) that uniquely identifies the destination backup vault for the copied backup.

