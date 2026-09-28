---
subcategory: "Timestream for InfluxDB"
layout: "aws"
page_title: "AWS: aws_timestreaminfluxdb_db_backups"
description: |-
  Terraform data source for listing Amazon Timestream for InfluxDB backups.
---

# Data Source: aws_timestreaminfluxdb_db_backups

Terraform data source for listing Amazon Timestream for InfluxDB backups.

## Example Usage

### List All Backups

```terraform
data "aws_timestreaminfluxdb_db_backups" "example" {}
```

### List Backups for a Specific Resource

```terraform
data "aws_timestreaminfluxdb_db_backups" "example" {
  db_resource_id = aws_timestreaminfluxdb_db_instance.example.id
}
```

## Argument Reference

The following arguments are optional:

* `db_resource_id` - (Optional) ID of the DB instance or DB cluster to list backups for. If omitted, all backups in the account and Region are returned.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `backups` - List of backups. See [`backups`](#backups) below.

### `backups`

* `arn` - ARN of the backup.
* `created_at` - Time when the backup was created, in RFC3339 format.
* `db_resource_id` - ID of the DB resource that the backup was created from.
* `deployment_type` - Deployment type of the resource that the backup was created from.
* `engine_type` - Engine type of the resource that the backup was created from.
* `expires_after` - Date after which the backup will be automatically deleted.
* `id` - ID of the backup.
* `kms_key_id` - AWS KMS key ARN used for encryption of the resource at the time of backup.
* `name` - Customer-provided name of the backup.
* `status` - Status of the backup.
* `type` - Type of backup.
