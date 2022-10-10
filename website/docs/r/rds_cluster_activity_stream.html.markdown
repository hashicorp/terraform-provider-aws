---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_activity_stream"
description: |-
  Manages RDS Aurora Cluster Database Activity Streams
---

# Resource: aws_rds_cluster_activity_stream

Manages RDS Aurora Cluster Database Activity Streams.

Database Activity Streams have some limits and requirements, refer to the [Monitoring Amazon Aurora using Database Activity Streams][1] documentation for detailed limitations and requirements.

~> **Note:** This resource always calls the RDS [`StartActivityStream`][2] API with the `ApplyImmediately` parameter set to `true`. This is because the Terraform needs the activity stream to be started in order for it to get the associated attributes.

~> **Note:** This resource depends on having at least one `aws_rds_cluster_instance` created. To avoid race conditions when all resources are being created together, add an explicit resource reference using the [resource `depends_on` meta-argument](/docs/configuration/resources.html#depends_on-explicit-resource-dependencies).

~> **Note:** This resource is available in all regions except the following: `cn-north-1`, `cn-northwest-1`, `us-gov-east-1`, `us-gov-west-1`

## Example Usage

```terraform
resource "aws_rds_cluster" "default" {
  cluster_identifier = "aurora-cluster-demo"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  engine             = "aurora-postgresql"
  engine_version     = "13.4"
}

resource "aws_rds_cluster_instance" "default" {
  identifier         = "aurora-instance-demo"
  cluster_identifier = aws_rds_cluster.default.cluster_identifier
  engine             = aws_rds_cluster.default.engine
  instance_class     = "db.r6g.large"
}

resource "aws_kms_key" "default" {
  description = "AWS KMS Key to encrypt Database Activity Stream"
}

resource "aws_rds_cluster_activity_stream" "default" {
  resource_arn = aws_rds_cluster.default.arn
  mode         = "async"
  kms_key_id   = aws_kms_key.default.key_id

  depends_on = [aws_rds_cluster_instance.default]
}
```


## Argument Reference

For more detailed documentation about each argument, refer to
the [AWS official documentation][3].

The following arguments are supported:

* `resource_arn` - (Required, Forces new resources) The Amazon Resource Name (ARN) of the DB cluster.
* `mode` - (Required, Forces new resources) Specifies the mode of the database activity stream. Database events such as a change or access generate an activity stream event. The database session can handle these events either synchronously or asynchronously. One of: `sync`, `async`.
* `kms_key_id` - (Required, Forces new resources) The AWS KMS key identifier for encrypting messages in the database activity stream. The AWS KMS key identifier is the key ARN, key ID, alias ARN, or alias name for the KMS key.
* `engine_native_audit_fields_included` - (Optional, Forces new resources) Specifies whether the database activity stream includes engine-native audit fields. This option only applies to an Oracle DB instance. By default, no engine-native audit fields are included. Defaults `false`.



## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the DB cluster.
* `kinesis_stream_name` - The name of the Amazon Kinesis data stream to be used for the database activity stream.


## Import

RDS Aurora Cluster Database Activity Streams can be imported using the `resource_arn`, e.g.

```
$ terraform import aws_rds_cluster_activity_stream.default arn:aws:rds:us-west-2:123456789012:cluster:aurora-cluster-demo
```

[1]: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/DBActivityStreams.html
[2]: https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_StartActivityStream.html
[3]: https://docs.aws.amazon.com/cli/latest/reference/rds/start-activity-stream.html
