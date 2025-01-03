---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_activity_stream"
description: |-
  Manages RDS Aurora Cluster Database Activity Streams
---

# Resource: aws_rds_cluster_activity_stream

Manages RDS Aurora Cluster Database Activity Streams.

Database Activity Streams have some limits and requirements, refer to the [Monitoring Amazon Aurora using Database Activity Streams][1] documentation and the [Supported Regions and Aurora DB engines for database activity streams][2] documentation for detailed limitations and requirements.

~> **Note:** This resource always calls the RDS [`StartActivityStream`][3] API with the `ApplyImmediately` parameter set to `true`. This is because the Terraform needs the activity stream to be started in order for it to get the associated attributes.

~> **Note:** This resource depends on having at least one `aws_rds_cluster_instance` created. To avoid race conditions when all resources are being created together, add an explicit resource reference using the [resource `depends_on` meta-argument](https://www.terraform.io/docs/configuration/resources.html#depends_on-explicit-resource-dependencies).

## Example Usage

```terraform
resource "aws_rds_cluster" "example" {
  cluster_identifier = "example-aurora-cluster"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "example"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  engine             = "aurora-postgresql"
  engine_version     = "13.4"
}

resource "aws_rds_cluster_instance" "example" {
  identifier         = "example-aurora-instance"
  cluster_identifier = aws_rds_cluster.example.cluster_identifier
  engine             = aws_rds_cluster.example.engine
  instance_class     = "db.r6g.large"
}

resource "aws_kms_key" "example" {
  description = "AWS KMS Key to encrypt Database Activity Stream"
}

resource "aws_rds_cluster_activity_stream" "example" {
  resource_arn = aws_rds_cluster.example.arn
  mode         = "async"
  kms_key_id   = aws_kms_key.example.key_id

  depends_on = [aws_rds_cluster_instance.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required, Forces new resources) Amazon Resource Name (ARN) of the DB cluster.
* `mode` - (Required, Forces new resources) Mode of the database activity stream. Database events such as a change or access generate an activity stream event. The database session can handle these events either synchronously or asynchronously. One of: `sync`, `async`.
* `kms_key_id` - (Required, Forces new resources) AWS KMS key identifier for encrypting messages in the database activity stream. The AWS KMS key identifier is the key ARN, key ID, alias ARN, or alias name for the KMS key.
* `engine_native_audit_fields_included` - (Optional, Forces new resources) Whether the database activity stream includes engine-native audit fields. This option applies to an Oracle or Microsoft SQL Server DB instance. By default, no engine-native audit fields are included. Defaults `false`.

  **Note:** Since this argument is not applicable to Aurora DB clusters, it should either not be set (which defaults to 'false') or be explicitly set to `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the DB cluster.
* `kinesis_stream_name` - Name of the Amazon Kinesis data stream to be used for the database activity stream.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS Aurora Cluster Database Activity Streams using the `resource_arn`. For example:

```terraform
import {
  to = aws_rds_cluster_activity_stream.default
  id = "arn:aws:rds:us-west-2:123456789012:cluster:aurora-cluster-demo"
}
```

Using `terraform import`, import RDS Aurora Cluster Database Activity Streams using the `resource_arn`. For example:

```console
% terraform import aws_rds_cluster_activity_stream.default arn:aws:rds:us-west-2:123456789012:cluster:aurora-cluster-demo
```

[1]: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/DBActivityStreams.html
[2]: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Concepts.Aurora_Fea_Regions_DB-eng.Feature.DBActivityStreams.html
[3]: https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_StartActivityStream.html
