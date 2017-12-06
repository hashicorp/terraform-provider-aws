---
layout: "aws"
page_title: "AWS: dynamodb_backup"
sidebar_current: "docs-aws-resource-dynamodb-backup"
description: |-
  Provides a DynamoDB backup resource
---

# aws\_dynamodb\_table

Provides a DynamoDB backup resource

## Example Usage

The following dynamodb table description models the table and GSI shown
in the [AWS SDK example documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.html)

```hcl
resource "aws_dynamodb_backup" "backup_example" {
  table_name = "table-1"
  backup_name = "backup-1"
}
```

## Argument Reference

The following arguments are supported:

* `table_name` - (Required) The name of the table.
* `backup_name` - (Required) The name of the backup.


## Attributes Reference

The following attributes are exported:

* `arn` - The arn of the backup
* `id` - The id of the backup
