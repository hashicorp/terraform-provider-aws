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

```terraform
resource "aws_sns_topic" "test" {
  name = "backup-vault-events"
}

data "aws_iam_policy_document" "test" {
  policy_id = "__default_policy_ID"

  statement {
    actions = [
      "SNS:Publish",
    ]

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["backup.amazonaws.com"]
    }

    resources = [
      aws_sns_topic.test.arn,
    ]

    sid = "__default_statement_ID"
  }
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_backup_vault_notifications" "test" {
  backup_vault_name   = "example_backup_vault"
  sns_topic_arn       = aws_sns_topic.test.arn
  backup_vault_events = ["BACKUP_JOB_STARTED", "RESTORE_JOB_COMPLETED"]
}
```

## Argument Reference

This resource supports the following arguments:

* `backup_vault_events` - (Required) Array of events that indicate the status of jobs to back up resources to the backup vault.
* `backup_vault_name` - (Required) Name of the backup vault to add notifications for.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sns_topic_arn` - (Required) ARN that specifies the topic for a backup vault’s events

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `backup_vault_arn` - ARN of the vault.
* `id` - Name of the vault.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup vault notifications using the `name`. For example:

```terraform
import {
  to = aws_backup_vault_notifications.test
  id = "TestVault"
}
```

Using `terraform import`, import Backup vault notifications using the `name`. For example:

```console
% terraform import aws_backup_vault_notifications.test TestVault
```
