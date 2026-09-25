---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_data_source"
description: |-
  Manages a Resource QuickSight Data Source.
---

# Resource: aws_quicksight_data_source

Resource for managing QuickSight Data Source

## Example Usage

### S3 Data Source

```terraform
resource "aws_quicksight_data_source" "default" {
  data_source_id = "example-id"
  name           = "My Cool Data in S3"

  parameters {
    s3 {
      manifest_file_location {
        bucket = "my-bucket"
        key    = "path/to/manifest.json"
      }
    }
  }

  type = "S3"
}
```

### S3 Data Source with IAM Role ARN

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "example" {
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket.example.bucket
  key    = "manifest.json"
  content = jsonencode({
    fileLocations = [
      {
        URIPrefixes = [
          "https://${aws_s3_bucket.example.id}.s3-${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
        ]
      }
    ]
    globalUploadSettings = {
      format         = "CSV"
      delimiter      = ","
      textqualifier  = "\""
      containsHeader = true
    }
  })
}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "quicksight.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_policy" "example" {
  name        = "example"
  description = "Policy to allow QuickSight access to S3 bucket"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action   = ["s3:GetObject"],
        Effect   = "Allow",
        Resource = "${aws_s3_bucket.example.arn}/${aws_s3_object.example.key}"
      },
      {
        Action   = ["s3:ListBucket"],
        Effect   = "Allow",
        Resource = aws_s3_bucket.example.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "example" {
  policy_arn = aws_iam_policy.example.arn
  role       = aws_iam_role.example.name
}

resource "aws_quicksight_data_source" "example" {
  data_source_id = "example-id"
  name           = "manifest in S3"

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.example.bucket
        key    = aws_s3_object.example.key
      }
      role_arn = aws_iam_role.example.arn
    }
  }

  type = "S3"
}
```

## Argument Reference

The following arguments are required:

* `data_source_id` - (Required, Forces new resource) Identifier for the data source.
* `name` - (Required) Name for the data source, maximum of 128 characters.
* `parameters` - (Required) Parameters used to connect to this data source (exactly one). See [`parameters` Block](#parameters-block) for details.
* `type` - (Required) Type of the data source. See the [AWS Documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CreateDataSource.html#QS-CreateDataSource-request-Type) for the complete list of valid values.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `credentials` - (Optional) Credentials Amazon QuickSight uses to connect to your underlying source. See [`credentials` Block](#credentials-block) below for more details.
* `permission` - (Optional) Set of resource permissions on the data source. Maximum of 64 items. See [`permission` Block](#permission-block) below for more details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `ssl_properties` - (Optional) Secure Socket Layer (SSL) properties that apply when Amazon QuickSight connects to your underlying source. See [`ssl_properties` Block](#ssl_properties-block) below for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_connection_properties` - (Optional) Use this parameter only when you want Amazon QuickSight to use a VPC connection when connecting to your underlying source. See [`vpc_connection_properties` Block](#vpc_connection_properties-block) below for more details.

### `credentials` Block

* `copy_source_arn` - (Optional, Conflicts with `credential_pair` and `secret_arn`) ARN of a data source that has the credential pair to use. When the value is not null, the `credential_pair` from the data source in the ARN is used.
* `credential_pair` - (Optional, Conflicts with `copy_source_arn` and `secret_arn`) Credential pair. See [`credential_pair` Block](#credential_pair-block) below for more details.
* `secret_arn` - (Optional, Conflicts with `copy_source_arn` and `credential_pair`) ARN of the secret associated with the data source in Amazon Secrets Manager.

### `credential_pair` Block

* `password` - (Required) Password, maximum length of 1024 characters.
* `username` - (Required) User name, maximum length of 64 characters.

### `parameters` Block

To specify data source connection parameters, exactly one of the following sub-objects must be provided.

* `amazon_elasticsearch` - (Optional) [Parameters](#amazon_elasticsearch-block) for connecting to Amazon Elasticsearch.
* `athena` - (Optional) [Parameters](#athena-block) for connecting to Athena.
* `aurora` - (Optional) [Parameters](#aurora-block) for connecting to Aurora MySQL.
* `aurora_postgresql` - (Optional) [Parameters](#aurora_postgresql-block) for connecting to Aurora Postgresql.
* `aws_iot_analytics` - (Optional) [Parameters](#aws_iot_analytics-block) for connecting to AWS IOT Analytics.
* `databricks` - (Optional) [Parameters](#databricks-block) for connecting to Databricks.
* `jira` - (Optional) [Parameters](#jira-block) for connecting to Jira.
* `maria_db` - (Optional) [Parameters](#maria_db-block) for connecting to MariaDB.
* `mysql` - (Optional) [Parameters](#mysql-block) for connecting to MySQL.
* `oracle` - (Optional) [Parameters](#oracle-block) for connecting to Oracle.
* `postgresql` - (Optional) [Parameters](#postgresql-block) for connecting to Postgresql.
* `presto` - (Optional) [Parameters](#presto-block) for connecting to Presto.
* `rds` - (Optional) [Parameters](#rds-block) for connecting to RDS.
* `redshift` - (Optional) [Parameters](#redshift-block) for connecting to Redshift.
* `s3` - (Optional) [Parameters](#s3-block) for connecting to S3.
* `service_now` - (Optional) [Parameters](#service_now-block) for connecting to ServiceNow.
* `snowflake` - (Optional) [Parameters](#snowflake-block) for connecting to Snowflake.
* `spark` - (Optional) [Parameters](#spark-block) for connecting to Spark.
* `sql_server` - (Optional) [Parameters](#sql_server-block) for connecting to SQL Server.
* `teradata` - (Optional) [Parameters](#teradata-block) for connecting to Teradata.
* `twitter` - (Optional) [Parameters](#twitter-block) for connecting to Twitter.

### `permission` Block

* `actions` - (Required) Set of IAM actions to grant or revoke permissions on. Max of 16 items.
* `principal` - (Required) ARN of the principal.

### `ssl_properties` Block

* `disable_ssl` - (Required) Whether to disable SSL.

### `vpc_connection_properties` Block

* `vpc_connection_arn` - (Required) ARN for the VPC connection.

### `amazon_elasticsearch` Block

* `domain` - (Required) OpenSearch domain.

### `athena` Block

* `role_arn` - (Optional) Use the `role_arn` to override an account-wide role for a specific Athena data source.
* `work_group` - (Optional) Work-group to which to connect.

### `aurora` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `aurora_postgresql` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `aws_iot_analytics` Block

* `data_set_name` - (Required) Name of the data set to which to connect.

### `databricks` Block

* `host` - (Required) Host name of the Databricks data source.
* `port` - (Required) Port for the Databricks data source.
* `sql_endpoint_path` - (Required) HTTP path of the Databricks data source.

### `jira` Block

* `site_base_url` - (Required) Base URL of the Jira instance's site to which to connect.

### `maria_db` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `mysql` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `oracle` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `postgresql` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `presto` Block

* `catalog` - (Required) Catalog to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `rds` Block

* `database` - (Required) Database to which to connect.
* `instance_id` - (Required) Instance ID to which to connect.

### `redshift` Block

* `cluster_id` - (Optional, Required if `host` and `port` are not provided) ID of the cluster to which to connect.
* `database` - (Required) Database to which to connect.
* `host` - (Optional, Required if `cluster_id` is not provided) Host to which to connect.
* `port` - (Optional, Required if `cluster_id` is not provided) Port to which to connect.

### `s3` Block

* `manifest_file_location` - (Required) S3 location of the S3 manifest file. See [`manifest_file_location` Block](#manifest_file_location-block) below for more details.
* `role_arn` - (Optional) Use the `role_arn` to override an account-wide role for a specific S3 data source. For example, say an account administrator has turned off all S3 access with an account-wide role. The administrator can then use `role_arn` to bypass the account-wide role and allow S3 access for the single S3 data source that is specified in the structure, even if the account-wide role forbidding S3 access is still active.

### `manifest_file_location` Block

* `bucket` - (Required) Name of the bucket that contains the manifest file.
* `key` - (Required) Key of the manifest file within the bucket.

### `service_now` Block

* `site_base_url` - (Required) Base URL of the ServiceNow instance's site to which to connect.

### `snowflake` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `warehouse` - (Required) Warehouse to which to connect.

### `spark` Block

* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `sql_server` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `teradata` Block

* `database` - (Required) Database to which to connect.
* `host` - (Required) Host to which to connect.
* `port` - (Required) Port to which to connect.

### `twitter` Block

* `max_rows` - (Required) Maximum number of rows to query.
* `query` - (Required) Twitter query to retrieve the data.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the data source
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight data source using the AWS account ID, and data source ID separated by a slash (`/`). For example:

```terraform
import {
  to = aws_quicksight_data_source.example
  id = "123456789123/my-data-source-id"
}
```

Using `terraform import`, import a QuickSight data source using the AWS account ID, and data source ID separated by a slash (`/`). For example:

```console
% terraform import aws_quicksight_data_source.example 123456789123/my-data-source-id
```
