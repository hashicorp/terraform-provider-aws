---
subcategory: "Timestream for InfluxDB"
layout: "aws"
page_title: "AWS: aws_timestreaminfluxdb_db_backup"
description: |-
  Terraform resource for managing an Amazon Timestream for InfluxDB on-demand backup.
---

# Resource: aws_timestreaminfluxdb_db_backup

Terraform resource for managing an Amazon Timestream for InfluxDB on-demand backup of a DB instance or DB cluster.

To configure recurring, automated backups instead, use the `db_backup_configuration` block on the [`aws_timestreaminfluxdb_db_instance`](timestreaminfluxdb_db_instance.html.markdown) or [`aws_timestreaminfluxdb_db_cluster`](timestreaminfluxdb_db_cluster.html.markdown) resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_timestreaminfluxdb_db_backup" "example" {
  db_resource_id = aws_timestreaminfluxdb_db_instance.example.id
  name           = "example-backup"
}
```

### Usage with Retention Period

```terraform
resource "aws_timestreaminfluxdb_db_backup" "example" {
  db_resource_id = aws_timestreaminfluxdb_db_instance.example.id
  name           = "example-backup"
  retention_days = 30
}
```

## Argument Reference

The following arguments are required:

* `db_resource_id` - (Required, Forces new resource) ID of the DB instance or DB cluster to back up.
* `name` - (Required, Forces new resource) Name of the backup. Must be unique within the account and Region.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `retention_days` - (Optional, Forces new resource) Number of days to retain the backup. Valid values are `1` to `3650`.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above. These attributes capture the configuration of the source resource at the time the backup was taken:

* `allocated_storage` - Allocated storage of the resource at the time of backup, in GiB.
* `arn` - ARN of the backup.
* `created_at` - Time when the backup was created, in RFC3339 format.
* `db_instance_type` - DB instance type of the resource at the time of backup.
* `db_parameter_group_id` - Identifier of the DB parameter group associated with the backup.
* `db_storage_type` - Storage type of the resource at the time of backup.
* `deployment_type` - Deployment type of the resource that the backup was created from.
* `engine_type` - Engine type of the resource that the backup was created from.
* `expires_after` - Date after which the backup will be automatically deleted.
* `failover_mode` - Failover mode of the resource at the time of backup.
* `id` - ID of the backup.
* `influx_auth_parameters_secret_arn` - ARN of the Secrets Manager secret containing the InfluxDB auth parameters.
* `kms_key_id` - AWS KMS key ARN used for encryption of the resource at the time of backup. When the source resource uses a customer managed key, the backup uses the same key.
* `network_type` - Network type of the resource at the time of backup.
* `port` - Port number of the resource at the time of backup.
* `publicly_accessible` - Whether the resource was publicly accessible at the time of backup.
* `status` - Current status of the backup.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `type` - Type of backup. For an on-demand backup this is `ON_DEMAND`.
* `vpc_security_group_ids` - VPC security group IDs associated with the resource at the time of backup.
* `vpc_subnet_ids` - VPC subnet IDs associated with the resource at the time of backup.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_timestreaminfluxdb_db_backup.example
  identity = {
    id = "01234567890abcdef"
  }
}

resource "aws_timestreaminfluxdb_db_backup" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the Timestream for InfluxDB backup.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Timestream for InfluxDB backups using `id`. For example:

```terraform
import {
  to = aws_timestreaminfluxdb_db_backup.example
  id = "01234567890abcdef"
}
```

Using `terraform import`, import Timestream for InfluxDB backups using `id`. For example:

```console
% terraform import aws_timestreaminfluxdb_db_backup.example 01234567890abcdef
```
