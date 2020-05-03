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
resource "aws_sns_topic" "topic" {
  name = "backup-events-notification"
}

resource "aws_backup_vault" "example" {
  name = "example_backup_vault"
}

resource "aws_backup_vault_notifications" "notify_team" {
  vault_name    = aws_backup_vault.example.name
  sns_topic_arn = aws_sns_topic.topic.arn

  events = [
    "BACKUP_JOB_STARTED",
    "BACKUP_JOB_COMPLETED",
    "BACKUP_JOB_SUCCESSFUL",
    "BACKUP_JOB_FAILED",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `vault_name`    - (Required) Name of the backup vault.
* `sns_topic_arn` - (Required) SNS Topic ARN to which event notifications are sent.
* `events`        - (Required) A list of events to listen to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the vault.
