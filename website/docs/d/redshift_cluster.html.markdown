---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_cluster"
description: |-
    Provides details about a specific redshift cluster
---

# Data Source: aws_redshift_cluster

Provides details about a specific redshift cluster.

## Example Usage

```terraform
data "aws_redshift_cluster" "example" {
  cluster_identifier = "example-cluster"
}

resource "aws_kinesis_firehose_delivery_stream" "example_stream" {
  name        = "terraform-kinesis-firehose-example-stream"
  destination = "redshift"

  redshift_configuration {
    role_arn           = aws_iam_role.firehose_role.arn
    cluster_jdbcurl    = "jdbc:redshift://${data.aws_redshift_cluster.example.endpoint}/${data.aws_redshift_cluster.example.database_name}"
    username           = "exampleuser"
    password           = "Exampl3Pass"
    data_table_name    = "example-table"
    copy_options       = "delimiter '|'" # the default delimiter
    data_table_columns = "example-col"

    s3_configuration {
      role_arn           = aws_iam_role.firehose_role.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffer_size        = 10
      buffer_interval    = 400
      compression_format = "GZIP"
    }
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_identifier` - (Required) Cluster identifier

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of cluster.
* `allow_version_upgrade` - Whether major version upgrades can be applied during maintenance period
* `automated_snapshot_retention_period` - The backup retention period
* `aqua_configuration_status` - The value represents how the cluster is configured to use AQUA.
* `availability_zone` - Availability zone of the cluster
* `availability_zone_relocation_enabled` - Indicates whether the cluster is able to be relocated to another availability zone.
* `bucket_name` - Name of the S3 bucket where the log files are to be stored
* `cluster_identifier` - Cluster identifier
* `cluster_nodes` - Nodes in the cluster. Cluster node blocks are documented below
* `cluster_parameter_group_name` - The name of the parameter group to be associated with this cluster
* `cluster_public_key` - Public key for the cluster
* `cluster_revision_number` - The cluster revision number
* `cluster_subnet_group_name` - The name of a cluster subnet group to be associated with this cluster
* `cluster_type` - Cluster type
* `cluster_namespace_arn` - The namespace Amazon Resource Name (ARN) of the cluster
* `database_name` - Name of the default database in the cluster
* `default_iam_role_arn` - The ARN for the IAM role that was set as default for the cluster when the cluster was created.
* `elastic_ip` - Elastic IP of the cluster
* `enable_logging` - Whether cluster logging is enabled
* `encrypted` - Whether the cluster data is encrypted
* `endpoint` - Cluster endpoint
* `enhanced_vpc_routing` - Whether enhanced VPC routing is enabled
* `iam_roles` - IAM roles associated to the cluster
* `kms_key_id` - KMS encryption key associated to the cluster
* `master_username` - Username for the master DB user
* `multi_az` - If the cluster is a Multi-AZ deployment
* `node_type` - Cluster node type
* `number_of_nodes` - Number of nodes in the cluster
* `maintenance_track_name` - The name of the maintenance track for the restored cluster.
* `manual_snapshot_retention_period` - (Optional)  The default number of days to retain a manual snapshot.
* `port` - Port the cluster responds on
* `preferred_maintenance_window` - The maintenance window
* `publicly_accessible` - Whether the cluster is publicly accessible
* `s3_key_prefix` - Folder inside the S3 bucket where the log files are stored
* `log_destination_type` - The log destination type.
* `log_exports` - Collection of exported log types. Log types include the connection log, user log and user activity log.
* `tags` - Tags associated to the cluster
* `vpc_id` - VPC Id associated with the cluster
* `vpc_security_group_ids` - The VPC security group Ids associated with the cluster

Cluster nodes (for `cluster_nodes`) support the following attributes:

* `node_role` - Whether the node is a leader node or a compute node
* `private_ip_address` - Private IP address of a node within a cluster
* `public_ip_address` - Public IP address of a node within a cluster
