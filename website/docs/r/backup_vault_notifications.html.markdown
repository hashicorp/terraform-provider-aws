---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault_notifications"
description: |-
  Provides an AWS Backup vault notifications resource.
---

# Resource: aws_backup_vault_notifications

Provides an AWS Backup vault notifications resource.

## Example Usage

```hcl
resource "aws_sns_topic" "test" {
  name = "backup-vault-events"
}

resource "aws_sns_topic_policy" "test" {
	arn = "${aws_sns_topic.test.arn}"
	policy = <<POLICY
{
      "Sid": "My-statement-id",
      "Effect": "Allow",
      "Principal": {
        "Service": "backup.amazonaws.com"
      },
      "Action": "SNS:Publish",
      "Resource": "${aws_sns_topic.test.arn}"
}
POLICY
}

resource "aws_backup_vault_notifications" "test" {
  backup_vault_name   = "example_backup_vault"
  sns_topic_arn       = "${sns_topic_arn.test.arn}"
  backup_vault_events = ["BACKUP_JOB_STARTED", "RESTORE_JOB_COMPLETED"] 
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the backup vault to add notifications for.
* `sns_topic_arn` - (Required) The Amazon Resource Name (ARN) that specifies the topic for a backup vault’s events
* `backup_vault_events` - (Required) An array of events that indicate the status of jobs to back up resources to the backup vault.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the vault.
* `arn` - The ARN of the vault.

## Import

Backup vault notifications can be imported using the `name`, e.g.

```
$ terraform import aws_backup_vault_notifications.test TestVault
```
