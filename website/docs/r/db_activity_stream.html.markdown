---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_activity_stream"
description: |-
  Manages RDS Database Activity Streams
---

# Resource: aws_db_activity_stream

Manages RDS Database Activity Streams.

Database Activity Streams have some limits and requirements, refer to the [Monitoring Amazon RDS with Database Activity Streams][1] documentation and the [Supported Regions and DB engines for database activity streams in Amazon RDS][2] documentation for detailed limitations and requirements.

~> **Note:** This resource always calls the RDS [`StartActivityStream`][3] API with the `ApplyImmediately` parameter set to `true`. This is because the Terraform needs the activity stream to be started in order for it to get the associated attributes.

## Example Usage

```terraform
resource "aws_kms_key" "example" {
  description = "AWS KMS Key to encrypt Database Activity Stream"
}

resource "aws_db_instance" "example" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = local.db_subnet_group_name
  engine                  = "sqlserver-se"
  engine_version          = "15.00"
  identifier              = "example-db-instance"
  instance_class          = "db.m6i.large"
  license_model           = "license-included"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  storage_encrypted       = true
  username                = "example"
}

resource "aws_db_activity_stream" "example" {
  resource_arn                        = aws_db_instance.example.arn
  kms_key_id                          = aws_kms_key.example.key_id
  mode                                = "async"
  engine_native_audit_fields_included = true
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required, Forces new resources) Amazon Resource Name (ARN) of the DB cluster.
* `mode` - (Required, Forces new resources) Mode of the database activity stream. Database events such as a change or access generate an activity stream event. The database session can handle these events either synchronously or asynchronously. One of: `sync`, `async`.
* `kms_key_id` - (Required, Forces new resources) AWS KMS key identifier for encrypting messages in the database activity stream. The AWS KMS key identifier is the key ARN, key ID, alias ARN, or alias name for the KMS key.
* `engine_native_audit_fields_included` - (Optional, Forces new resources) Whether the database activity stream includes engine-native audit fields. This option applies to an Oracle or Microsoft SQL Server DB instance. By default, no engine-native audit fields are included. Defaults `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the DB cluster.
* `kinesis_stream_name` - Name of the Amazon Kinesis data stream to be used for the database activity stream.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS Database Activity Streams using the `resource_arn`. For example:

```terraform
import {
  to = aws_db_activity_stream.default
  id = "arn:aws:rds:us-west-2:123456789012:db:my-mysql-instance-1"
}
```

Using `terraform import`, import RDS Database Activity Streams using the `resource_arn`. For example:

```console
% terraform import aws_db_activity_stream.default arn:aws:rds:us-west-2:123456789012:db:my-mysql-instance-1
```

[1]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/DBActivityStreams.html
[2]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RDS_Fea_Regions_DB-eng.Feature.DBActivityStreams.html
[3]: https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_StartActivityStream.html
