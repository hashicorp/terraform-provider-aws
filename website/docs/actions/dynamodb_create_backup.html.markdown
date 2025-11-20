---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_create_backup"
description: |-
  Creates an on-demand backup of a DynamoDB table.
---

# Action: aws_dynamodb_create_backup

~> **Note:** `aws_dynamodb_create_backup` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Creates an on-demand backup of a DynamoDB table. This action will initiate a backup and wait for it to complete, providing progress updates during execution.

For information about DynamoDB backups, see the [DynamoDB Developer Guide](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/BackupRestore.html). For specific information about creating backups, see the [CreateBackup](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateBackup.html) page in the DynamoDB API Reference.

~> **Note:** On-demand backups do not consume provisioned throughput and have no impact on table performance.

## Example Usage

### Basic Usage

```terraform
resource "aws_dynamodb_table" "example" {
  name         = "example-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

action "aws_dynamodb_create_backup" "example" {
  config {
    table_name = aws_dynamodb_table.example.name
  }
}

resource "terraform_data" "backup_trigger" {
  input = "trigger-backup"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_dynamodb_create_backup.example]
    }
  }
}
```

### Backup with Custom Name

```terraform
action "aws_dynamodb_create_backup" "named" {
  config {
    table_name  = aws_dynamodb_table.production.name
    backup_name = "production-backup-${formatdate("YYYY-MM-DD", timestamp())}"
  }
}
```

## Argument Reference

The following arguments are required:

* `table_name` - (Required) Name or ARN of the DynamoDB table to backup. Must be between 1 and 1024 characters.

The following arguments are optional:

* `backup_name` - (Optional) Name for the backup. If not provided, a unique name will be generated automatically using the table name and a unique identifier. Must be between 3 and 255 characters and contain only alphanumeric characters, underscores, periods, and hyphens.
* `timeout` - (Optional) Timeout in minutes for the backup operation. Defaults to 10 minutes.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
