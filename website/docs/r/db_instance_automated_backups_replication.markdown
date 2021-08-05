---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_instance_automated_backups_replication"
description: |-
  Manages an RDS DB Instance Automated Backups Replication.
---

# Resource: aws_db_instance_automated_backups_replication

Manages an RDS DB Instance Automated Backups Replication.

~> **NOTE:** This resource requires a second AWS provider to be defined in another region.

* [Replicating automated backups to another AWS Region](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ReplicateBackups.html)

## Example Usage

```terraform
resource "aws_db_instance_automated_backups_replication" "example" {
  source_db_instance_arn = aws_db_instance.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `backup_retention_period` - (Optional) The retention period for the replicated automated backups.
* `kms_key_id` - (Optional) The AWS KMS key identifier for encryption of the replicated automated backups.
* `pre_signed_url` - (Optional) A URL that contains a Signature Version 4 signed request for the [StartDBInstanceAutomatedBackupsReplication](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_StartDBInstanceAutomatedBackupsReplication.html) action to be called in the AWS Region of the source DB instance.
* `source_db_instance_arn` - (Optional) The Amazon Resource Name (ARN) of the source DB instance for the replicated automated backups.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - DB Instance Automated Backup Replication ARN

## Import

`aws_db_instance_automated_backups_replication` can be imported using the DB Instance Automated Backup Replication ARN, e.g.

```
$ terraform import aws_db_instance_automated_backups_replication.example arn:aws:rds:eu-west-1:123456789012:auto-backup:ab-lrg8qb6qtarwcfvoto3so53igbdulp3xjs8xeym
```
