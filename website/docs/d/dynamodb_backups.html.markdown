---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_backups"
description: |-
  Data source for listing AWS DynamoDB backups.
---

# Data Source: aws_dynamodb_backups

Data source for listing AWS DynamoDB backups.

## Example Usage

### Basic Usage

```terraform
data "aws_dynamodb_backups" "example" {
  table_name = "my-table"
}
```

## Argument Reference

This data source supports the following arguments:

* `backup_type` - (Optional) Backup type. Valid values: `USER`, `SYSTEM`, `AWS_BACKUP`, `ALL`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `table_name` - (Optional) Name of the table to list backups for.
* `time_range_lower_bound` - (Optional) Only backups created after this time are listed. Time must be in RFC3339 format.
* `time_range_upper_bound` - (Optional) Only backups created before this time are listed. Time must be in RFC3339 format.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `backup_summaries` - List of backups. See below.

### `backup_summaries` Attribute Reference

* `backup_arn` - ARN of the backup.
* `backup_creation_date_time` - Time at which the backup was created.
* `backup_expiry_date_time` - Time at which the automatic on-demand backup created by DynamoDB will expire.
* `backup_name` - Name of the specified backup.
* `backup_size_bytes` - Size of the backup in bytes.
* `backup_status` - Backup can be in one of the following states: `CREATING`, `DELETED`, `AVAILABLE`.
* `backup_type` - BackupType: `USER`, `SYSTEM`, `AWS_BACKUP`.
* `table_arn` - ARN associated with the table.
* `table_id` - Unique identifier for the table.
* `table_name` - Name of the table.
