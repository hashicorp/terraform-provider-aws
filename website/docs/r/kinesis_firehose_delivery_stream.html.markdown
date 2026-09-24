---
subcategory: "Kinesis Firehose"
layout: "aws"
page_title: "AWS: aws_kinesis_firehose_delivery_stream"
description: |-
  Provides a AWS Kinesis Firehose Delivery Stream
---

# Resource: aws_kinesis_firehose_delivery_stream

Provides a Kinesis Firehose Delivery Stream resource. Amazon Kinesis Firehose is a fully managed, elastic service to easily deliver real-time data streams to destinations such as Amazon S3 , Amazon Redshift and Snowflake.

For more details, see the [Amazon Kinesis Firehose Documentation](https://aws.amazon.com/documentation/firehose/).

## Example Usage

### Extended S3 Destination

```terraform
resource "aws_kinesis_firehose_delivery_stream" "extended_s3_stream" {
  name        = "terraform-kinesis-firehose-extended-s3-test-stream"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose_role.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    processing_configuration {
      enabled = "true"

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_processor.arn}:$LATEST"
        }
      }
    }
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket"
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

data "aws_iam_policy_document" "firehose_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["firehose.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "firehose_role" {
  name               = "firehose_test_role"
  assume_role_policy = data.aws_iam_policy_document.firehose_assume_role.json
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "lambda_iam" {
  name               = "lambda_iam"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_lambda_function" "lambda_processor" {
  filename      = "lambda.zip"
  function_name = "firehose_lambda_processor"
  role          = aws_iam_role.lambda_iam.arn
  handler       = "exports.handler"
  runtime       = "nodejs24.x"
}
```

### Extended S3 Destination with dynamic partitioning

These examples use built-in Firehose functionality, rather than requiring a lambda.

```terraform
resource "aws_kinesis_firehose_delivery_stream" "extended_s3_stream" {
  name        = "terraform-kinesis-firehose-extended-s3-test-stream"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose_role.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    buffering_size = 64

    # https://docs.aws.amazon.com/firehose/latest/dev/dynamic-partitioning.html
    dynamic_partitioning_configuration {
      enabled = "true"
    }

    # Example prefix using partitionKeyFromQuery, applicable to JQ processor
    prefix              = "data/customer_id=!{partitionKeyFromQuery:customer_id}/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/"
    error_output_prefix = "errors/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/!{firehose:error-output-type}/"

    processing_configuration {
      enabled = "true"

      # Multi-record deaggregation processor example
      processors {
        type = "RecordDeAggregation"
        parameters {
          parameter_name  = "SubRecordType"
          parameter_value = "JSON"
        }
      }

      # New line delimiter processor example
      processors {
        type = "AppendDelimiterToRecord"
      }

      # JQ processor example
      processors {
        type = "MetadataExtraction"
        parameters {
          parameter_name  = "JsonParsingEngine"
          parameter_value = "JQ-1.6"
        }
        parameters {
          parameter_name  = "MetadataExtractionQuery"
          parameter_value = "{customer_id:.customer_id}"
        }
      }
    }
  }
}
```

Multiple Dynamic Partitioning Keys (maximum of 50) can be added by comma separating the `parameter_value`.

The following example adds the Dynamic Partitioning Keys: `store_id` and `customer_id` to the S3 prefix.

```terraform
resource "aws_kinesis_firehose_delivery_stream" "extended_s3_stream" {
  name        = "terraform-kinesis-firehose-extended-s3-test-stream"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn       = aws_iam_role.firehose_role.arn
    bucket_arn     = aws_s3_bucket.bucket.arn
    buffering_size = 64
    # https://docs.aws.amazon.com/firehose/latest/dev/dynamic-partitioning.html
    dynamic_partitioning_configuration {
      enabled = "true"
    }
    # Example prefix using partitionKeyFromQuery, applicable to JQ processor
    prefix              = "data/store_id=!{partitionKeyFromQuery:store_id}/customer_id=!{partitionKeyFromQuery:customer_id}/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/"
    error_output_prefix = "errors/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/!{firehose:error-output-type}/"
    processing_configuration {
      enabled = "true"
      # JQ processor example
      processors {
        type = "MetadataExtraction"
        parameters {
          parameter_name  = "JsonParsingEngine"
          parameter_value = "JQ-1.6"
        }
        parameters {
          parameter_name  = "MetadataExtractionQuery"
          parameter_value = "{store_id:.store_id,customer_id:.customer_id}"
        }
      }
    }
  }
}
```

### Redshift Destination

```terraform
resource "aws_redshift_cluster" "test_cluster" {
  cluster_identifier = "tf-redshift-cluster"
  database_name      = "test"
  master_username    = "testuser"
  master_password    = "T3stPass"
  node_type          = "dc1.large"
  cluster_type       = "single-node"
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "redshift"

  redshift_configuration {
    role_arn           = aws_iam_role.firehose_role.arn
    cluster_jdbcurl    = "jdbc:redshift://${aws_redshift_cluster.test_cluster.endpoint}/${aws_redshift_cluster.test_cluster.database_name}"
    username           = "testuser"
    password           = "T3stPass"
    data_table_name    = "test-table"
    copy_options       = "delimiter '|'" # the default delimiter
    data_table_columns = "test-col"
    s3_backup_mode     = "Enabled"

    s3_configuration {
      role_arn           = aws_iam_role.firehose_role.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }

    s3_backup_configuration {
      role_arn           = aws_iam_role.firehose_role.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 15
      buffering_interval = 300
      compression_format = "GZIP"
    }
  }
}
```

### Elasticsearch Destination

```terraform
resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = "firehose-es-test"
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "elasticsearch"

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose_role.arn
    index_name = "test"
    type_name  = "test"

    s3_configuration {
      role_arn           = aws_iam_role.firehose_role.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }

    processing_configuration {
      enabled = "true"

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_processor.arn}:$LATEST"
        }
      }
    }
  }
}
```

### Elasticsearch Destination With VPC

```terraform
resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = "es-test"

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  vpc_options {
    security_group_ids = [aws_security_group.first.id]
    subnet_ids         = [aws_subnet.first.id, aws_subnet.second.id]
  }
}

data "aws_iam_policy_document" "firehose-elasticsearch" {
  statement {
    effect  = "Allow"
    actions = ["es:*"]

    resources = [
      aws_elasticsearch_domain.test_cluster.arn,
      "${aws_elasticsearch_domain.test_cluster.arn}/*",
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "ec2:DescribeVpcs",
      "ec2:DescribeVpcAttribute",
      "ec2:DescribeSubnets",
      "ec2:DescribeSecurityGroups",
      "ec2:DescribeNetworkInterfaces",
      "ec2:CreateNetworkInterface",
      "ec2:CreateNetworkInterfacePermission",
      "ec2:DeleteNetworkInterface",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "firehose-elasticsearch" {
  name   = "elasticsearch"
  role   = aws_iam_role.firehose.id
  policy = data.aws_iam_policy_document.firehose-elasticsearch.json
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = "terraform-kinesis-firehose-es"
  destination = "elasticsearch"

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    vpc_config {
      subnet_ids         = [aws_subnet.first.id, aws_subnet.second.id]
      security_group_ids = [aws_security_group.first.id]
      role_arn           = aws_iam_role.firehose.arn
    }
  }
}
```

### OpenSearch Destination

```terraform
resource "aws_opensearch_domain" "test_cluster" {
  domain_name = "firehose-os-test"
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "opensearch"

  opensearch_configuration {
    domain_arn = aws_opensearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose_role.arn
    index_name = "test"

    s3_configuration {
      role_arn           = aws_iam_role.firehose_role.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }

    processing_configuration {
      enabled = "true"

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_processor.arn}:$LATEST"
        }
      }
    }
  }
}
```

### OpenSearch Destination With VPC

```terraform
resource "aws_opensearch_domain" "test_cluster" {
  domain_name = "es-test"

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "m4.large.search"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  vpc_options {
    security_group_ids = [aws_security_group.first.id]
    subnet_ids         = [aws_subnet.first.id, aws_subnet.second.id]
  }
}

resource "aws_iam_role_policy" "firehose-opensearch" {
  name   = "opensearch"
  role   = aws_iam_role.firehose.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "es:*"
      ],
      "Resource": [
        "${aws_opensearch_domain.test_cluster.arn}",
        "${aws_opensearch_domain.test_cluster.arn}/*"
      ]
        },
        {
          "Effect": "Allow",
          "Action": [
            "ec2:DescribeVpcs",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeSubnets",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeNetworkInterfaces",
            "ec2:CreateNetworkInterface",
            "ec2:CreateNetworkInterfacePermission",
            "ec2:DeleteNetworkInterface"
          ],
          "Resource": [
            "*"
          ]
        }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-opensearch]

  name        = "terraform-kinesis-firehose-os"
  destination = "opensearch"

  opensearch_configuration {
    domain_arn = aws_opensearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    vpc_config {
      subnet_ids         = [aws_subnet.first.id, aws_subnet.second.id]
      security_group_ids = [aws_security_group.first.id]
      role_arn           = aws_iam_role.firehose.arn
    }
  }
}
```

### OpenSearch Serverless Destination

```terraform
resource "aws_opensearchserverless_collection" "test_collection" {
  name = "firehose-osserverless-test"
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "opensearchserverless"

  opensearchserverless_configuration {
    collection_endpoint = aws_opensearchserverless_collection.test_collection.collection_endpoint
    role_arn            = aws_iam_role.firehose_role.arn
    index_name          = "test"

    s3_configuration {
      role_arn           = aws_iam_role.firehose_role.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }

    processing_configuration {
      enabled = "true"

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_processor.arn}:$LATEST"
        }
      }
    }
  }
}
```

### Iceberg Destination

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "test-bucket"
  force_destroy = true
}

resource "aws_glue_catalog_database" "test" {
  name = "test"
}

resource "aws_glue_catalog_table" "test" {
  name          = "test"
  database_name = aws_glue_catalog_database.test.name
  parameters = {
    format = "parquet"
  }

  table_type = "EXTERNAL_TABLE"

  open_table_format_input {
    iceberg_input {
      metadata_operation = "CREATE"
      version            = 2
    }
  }

  storage_descriptor {
    location = "s3://${aws_s3_bucket.bucket.id}"

    columns {
      name = "my_column_1"
      type = "int"
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "iceberg"

  iceberg_configuration {
    role_arn           = aws_iam_role.firehose_role.arn
    catalog_arn        = "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:catalog"
    buffering_size     = 10
    buffering_interval = 400

    s3_configuration {
      role_arn   = aws_iam_role.firehose_role.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    destination_table_configuration {
      database_name = aws_glue_catalog_database.test.name
      table_name    = aws_glue_catalog_table.test.name
    }

    processing_configuration {
      enabled = "true"

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_processor.arn}:$LATEST"
        }
      }
    }
  }
}
```

### Splunk Destination

```terraform
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "splunk"

  splunk_configuration {
    hec_endpoint               = "https://http-inputs-mydomain.splunkcloud.com:443"
    hec_token                  = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"
    hec_acknowledgment_timeout = 600
    hec_endpoint_type          = "Event"
    s3_backup_mode             = "FailedEventsOnly"

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }
  }
}
```

### HTTP Endpoint (e.g., New Relic, Datadog) Destination

The HTTP endpoint destination can be used with any vendor that exposes a compatible HTTP intake, including [New Relic](https://docs.newrelic.com/docs/infrastructure/amazon-integrations/connect/aws-firehose/) and [Datadog](https://docs.datadoghq.com/integrations/amazon_kinesis_data_firehose/). Refer to each vendor's documentation for the correct intake URL and authentication requirements.

```terraform
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "http_endpoint"

  http_endpoint_configuration {
    url                = "https://aws-api.newrelic.com/firehose/v1"
    name               = "New Relic"
    access_key         = "my-key"
    buffering_size     = 15
    buffering_interval = 600
    role_arn           = aws_iam_role.firehose.arn
    s3_backup_mode     = "FailedDataOnly"

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }

    request_configuration {
      content_encoding = "GZIP"

      common_attributes {
        name  = "testname"
        value = "testvalue"
      }

      common_attributes {
        name  = "testname2"
        value = "testvalue2"
      }
    }
  }
}
```

### Snowflake Destination

```terraform
resource "aws_kinesis_firehose_delivery_stream" "example_snowflake_destination" {
  name        = "example-snowflake-destination"
  destination = "snowflake"

  snowflake_configuration {
    account_url        = "https://example.snowflakecomputing.com"
    buffering_size     = 15
    buffering_interval = 600
    database           = "example-db"
    private_key        = "..."
    role_arn           = aws_iam_role.firehose.arn
    schema             = "example-schema"
    table              = "example-table"
    user               = "example-usr"

    s3_configuration {
      role_arn           = aws_iam_role.firehose.arn
      bucket_arn         = aws_s3_bucket.bucket.arn
      buffering_size     = 10
      buffering_interval = 400
      compression_format = "GZIP"
    }
  }
}
```

### Data Format Conversion

```terraform
resource "aws_kinesis_firehose_delivery_stream" "example" {
  # ... other configuration ...
  extended_s3_configuration {
    # Must be at least 64
    buffering_size = 128

    # ... other configuration ...
    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {}
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.example.database_name
        role_arn      = aws_iam_role.example.arn
        table_name    = aws_glue_catalog_table.example.name
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `destination` - (Required) Destination where the data is delivered. The only options are `s3` (Deprecated, use `extended_s3` instead), `extended_s3`, `redshift`, `elasticsearch`, `splunk`, `http_endpoint`, `opensearch`, `opensearchserverless` and `snowflake`.
* `destination_id` - (Optional) Identifier of the destination for the delivery stream.
* `elasticsearch_configuration` - (Optional) Configuration options when `destination` is `elasticsearch`. See [`elasticsearch_configuration` block](#elasticsearch_configuration-block) below for details.
* `extended_s3_configuration` - (Optional, only Required when `destination` is `extended_s3`) Enhanced configuration options for the s3 destination. See [`extended_s3_configuration` block](#extended_s3_configuration-block) below for details.
* `http_endpoint_configuration` - (Optional) Configuration options when `destination` is `http_endpoint`. Requires the user to also specify an `s3_configuration` block.  See [`http_endpoint_configuration` block](#http_endpoint_configuration-block) below for details.
* `iceberg_configuration` - (Optional) Configuration options when `destination` is `iceberg`. See [`iceberg_configuration` block](#iceberg_configuration-block) below for details.
* `kinesis_source_configuration` - (Optional) Stream and role ARNs for a Kinesis data stream used as the source for a delivery stream. See [`kinesis_source_configuration` block](#kinesis_source_configuration-block) below for details.
* `msk_source_configuration` - (Optional) Configuration for the Amazon MSK cluster to be used as the source for a delivery stream. See [`msk_source_configuration` block](#msk_source_configuration-block) below for details.
* `name` - (Required) Name to identify the stream. This is unique to the AWS account and region the Stream is created in. When using for WAF logging, name must be prefixed with `aws-waf-logs-`. See [AWS Documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-policies.html#waf-policies-logging-config) for more details.
* `opensearch_configuration` - (Optional) Configuration options when `destination` is `opensearch`. See [`opensearch_configuration` block](#opensearch_configuration-block) below for details.
* `opensearchserverless_configuration` - (Optional) Configuration options when `destination` is `opensearchserverless`. See [`opensearchserverless_configuration` block](#opensearchserverless_configuration-block) below for details.
* `redshift_configuration` - (Optional) Configuration options when `destination` is `redshift`. Requires the user to also specify an `s3_configuration` block. See [`redshift_configuration` block](#redshift_configuration-block) below for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `server_side_encryption` - (Optional) Encrypt at rest options. See [`server_side_encryption` block](#server_side_encryption-block) below for details.
* `snowflake_configuration` - (Optional) Configuration options when `destination` is `snowflake`. See [`snowflake_configuration` block](#snowflake_configuration-block) below for details.
* `splunk_configuration` - (Optional) Configuration options when `destination` is `splunk`. See [`splunk_configuration` block](#splunk_configuration-block) below for details.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `version_id` - (Optional) Version identifier of the delivery stream.

~> **NOTE:** Server-side encryption should not be enabled when a kinesis stream is configured as the source of the firehose delivery stream.

### `kinesis_source_configuration` Block

The `kinesis_source_configuration` configuration block supports the following arguments:

* `kinesis_stream_arn` - (Required) Kinesis stream used as the source of the firehose delivery stream.
* `role_arn` - (Required) ARN of the role that provides access to the source Kinesis stream.

### `msk_source_configuration` Block

The `msk_source_configuration` configuration block supports the following arguments:

* `authentication_configuration` - (Required) Authentication configuration of the Amazon MSK cluster. See [`authentication_configuration` block](#authentication_configuration-block) below for details.
* `msk_cluster_arn` - (Required) ARN of the Amazon MSK cluster.
* `read_from_timestamp` - (Optional) Start date and time in UTC for the offset position within your MSK topic from where Firehose begins to read. By default, this is set to timestamp when Firehose becomes Active. If you want to create a Firehose stream with Earliest start position set the `read_from_timestamp` parameter to Epoch (1970-01-01T00:00:00Z).
* `topic_name` - (Required) Topic name within the Amazon MSK cluster.

### `authentication_configuration` Block

The `authentication_configuration` configuration block supports the following arguments:

* `connectivity` - (Required) Type of connectivity used to access the Amazon MSK cluster. Valid values: `PUBLIC`, `PRIVATE`.
* `role_arn` - (Required) ARN of the role used to access the Amazon MSK cluster.

### `server_side_encryption` Block

The `server_side_encryption` configuration block supports the following arguments:

* `enabled` - (Optional) Whether to enable encryption at rest. Default is `false`.
* `key_arn` - (Optional) ARN of the encryption key. Required when `key_type` is `CUSTOMER_MANAGED_CMK`.
* `key_type` - (Optional) Type of encryption key. Default is `AWS_OWNED_CMK`. Valid values are `AWS_OWNED_CMK` and `CUSTOMER_MANAGED_CMK`.

### `extended_s3_configuration` Block

The `extended_s3_configuration` configuration block supports the following arguments:

* `bucket_arn` - (Required) ARN of the S3 bucket.
* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds, before delivering it to the destination. The default value is 300.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs, before delivering it to the destination. The default value is 5. We recommend setting this value greater than the amount of data you typically ingest into the delivery stream in 10 seconds. For example, if you typically ingest data at 1 MB/sec, set this value to 10 MB or higher.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `compression_format` - (Optional) Compression format. If no value is specified, the default is `UNCOMPRESSED`. Other supported values are `GZIP`, `ZIP`, `Snappy`, & `HADOOP_SNAPPY`.
* `custom_time_zone` - (Optional) Time zone you prefer. Valid values are `UTC` or a non-3-letter IANA time zones (for example, `America/Los_Angeles`). Default value is `UTC`.
* `data_format_conversion_configuration` - (Optional) Nested argument for the serializer, deserializer, and schema for converting data from the JSON format to the Parquet or ORC format before writing it to Amazon S3. See [`data_format_conversion_configuration` block](#data_format_conversion_configuration-block) below for details.
* `dynamic_partitioning_configuration` - (Optional) Configuration for dynamic partitioning. Required when using [dynamic partitioning](https://docs.aws.amazon.com/firehose/latest/dev/dynamic-partitioning.html). See [`dynamic_partitioning_configuration` block](#dynamic_partitioning_configuration-block) below for details.
* `error_output_prefix` - (Optional) Prefix added to failed records before writing them to S3. Not currently supported for `redshift` destination. This prefix appears immediately following the bucket name. For information about how to specify this prefix, see [Custom Prefixes for Amazon S3 Objects](https://docs.aws.amazon.com/firehose/latest/dev/s3-prefixes.html).
* `file_extension` - (Optional) File extension to override the default file extension (for example, `.json`).
* `kms_key_arn` - (Optional) KMS key ARN the stream will use to encrypt data. If not set, no encryption will be used.
* `prefix` - (Optional) Extra prefix added in front of the "YYYY/MM/DD/HH" time format prefix that is automatically used for delivered S3 files. Note that if the prefix ends with a slash, it appears as a folder in the S3 bucket.
* `processing_configuration` - (Optional) Data processing configuration. See [`processing_configuration` block](#processing_configuration-block) below for details.
* `role_arn` - (Required) ARN of the AWS credentials.
* `s3_backup_configuration` - (Optional) Configuration for backup in Amazon S3. Required if `s3_backup_mode` is `Enabled`. See [`s3_backup_configuration` block](#s3_backup_configuration-block) below for details.
* `s3_backup_mode` - (Optional) Amazon S3 backup mode. Valid values are `Disabled` and `Enabled`. Default value is `Disabled`.

### `s3_backup_configuration` Block

The `s3_backup_configuration` configuration block supports the following arguments:

* `bucket_arn` - (Required) ARN of the S3 bucket.
* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds, before delivering it to the destination. The default value is 300.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs, before delivering it to the destination. The default value is 5.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `compression_format` - (Optional) Compression format. If no value is specified, the default is `UNCOMPRESSED`. Other supported values are `GZIP`, `ZIP`, `Snappy`, & `HADOOP_SNAPPY`.
* `error_output_prefix` - (Optional) Prefix added to failed records before writing them to S3. This prefix appears immediately following the bucket name. For information about how to specify this prefix, see [Custom Prefixes for Amazon S3 Objects](https://docs.aws.amazon.com/firehose/latest/dev/s3-prefixes.html).
* `kms_key_arn` - (Optional) KMS key ARN the stream will use to encrypt data. If not set, no encryption will be used.
* `prefix` - (Optional) Extra prefix added in front of the "YYYY/MM/DD/HH" time format prefix that is automatically used for delivered S3 files. Note that if the prefix ends with a slash, it appears as a folder in the S3 bucket.
* `role_arn` - (Required) ARN of the AWS credentials.

### `redshift_configuration` Block

The `redshift_configuration` configuration block supports the following arguments:

* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `cluster_jdbcurl` - (Required) JDBC URL of the redshift cluster.
* `copy_options` - (Optional) Copy options for copying the data from the s3 intermediate bucket into redshift, for example to change the default delimiter. For valid values, see the [AWS documentation](http://docs.aws.amazon.com/firehose/latest/APIReference/API_CopyCommand.html).
* `data_table_columns` - (Optional) Data table columns that will be targeted by the copy command.
* `data_table_name` - (Required) Name of the table in the redshift cluster that the s3 bucket will copy to.
* `password` - (Optional) Password for the username above. This value is required if `secrets_manager_configuration` is not provided.
* `processing_configuration` - (Optional) Data processing configuration. See [`processing_configuration` block](#processing_configuration-block) below for details.
* `retry_duration` - (Optional) Length of time during which Firehose retries delivery after a failure, starting from the initial request and including the first attempt. The default value is 3600 seconds (60 minutes). Firehose does not retry if the value of DurationInSeconds is 0 (zero) or if the first delivery attempt takes longer than the current value.
* `role_arn` - (Required) ARN of the role the stream assumes.
* `s3_backup_configuration` - (Optional) Configuration for backup in Amazon S3. Required if `s3_backup_mode` is `Enabled`. See [`s3_backup_configuration` block](#s3_backup_configuration-block) below for details.
* `s3_backup_mode` - (Optional) Amazon S3 backup mode. Valid values are `Disabled` and `Enabled`. Default value is `Disabled`.
* `s3_configuration` - (Required) S3 Configuration. See [`s3_configuration` block](#s3_configuration-block) below for details.
* `secrets_manager_configuration` - (Optional) Secrets Manager configuration. See [`secrets_manager_configuration` block](#secrets_manager_configuration-block) below for details. This value is required if `username` and `password` are not provided.
* `username` - (Optional) Username that the firehose delivery stream will assume. It is strongly recommended that the username and password provided is used exclusively for Amazon Kinesis Firehose purposes, and that the permissions for the account are restricted for Amazon Redshift INSERT permissions. This value is required if `secrets_manager_configuration` is not provided.

### `elasticsearch_configuration` Block

The `elasticsearch_configuration` configuration block supports the following arguments:

* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds between 0 to 900, before delivering it to the destination.  The default value is 300s.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs between 1 to 100, before delivering it to the destination.  The default value is 5MB.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `cluster_endpoint` - (Optional) Endpoint to use when communicating with the cluster. Conflicts with `domain_arn`.
* `domain_arn` - (Optional) ARN of the Amazon ES domain.  The pattern needs to be `arn:.*`.  Conflicts with `cluster_endpoint`.
* `index_name` - (Required) Elasticsearch index name.
* `index_rotation_period` - (Optional) Elasticsearch index rotation period.  Index rotation appends a timestamp to the IndexName to facilitate expiration of old data.  Valid values are `NoRotation`, `OneHour`, `OneDay`, `OneWeek`, and `OneMonth`.  The default value is `OneDay`.
* `processing_configuration` - (Optional) Data processing configuration.  See [`processing_configuration` block](#processing_configuration-block) below for details.
* `retry_duration` - (Optional) After an initial failure to deliver to Amazon Elasticsearch, the total amount of time, in seconds between 0 to 7200, during which Firehose re-attempts delivery (including the first attempt).  After this time has elapsed, the failed documents are written to Amazon S3.  The default value is 300s.  There will be no retry if the value is 0.
* `role_arn` - (Required) ARN of the IAM role to be assumed by Firehose for calling the Amazon ES Configuration API and for indexing documents.  The IAM role must have permission for `DescribeElasticsearchDomain`, `DescribeElasticsearchDomains`, and `DescribeElasticsearchDomainConfig`.  The pattern needs to be `arn:.*`.
* `s3_backup_mode` - (Optional) How documents should be delivered to Amazon S3.  Valid values are `FailedDocumentsOnly` and `AllDocuments`.  Default value is `FailedDocumentsOnly`.
* `s3_configuration` - (Required) S3 Configuration. See [`s3_configuration` block](#s3_configuration-block) below for details.
* `type_name` - (Optional) Elasticsearch type name with maximum length of 100 characters.
* `vpc_config` - (Optional) VPC configuration for the delivery stream to connect to Elastic Search associated with the VPC. See [`vpc_config` block](#vpc_config-block) below for details.

### `iceberg_configuration` Block

The `iceberg_configuration` configuration block supports the following arguments:

* `append_only` - (Optional) Whether to enable append only operations. When set, Firehose only performs insert operations on the destination Apache Iceberg Tables.
* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds between 0 and 900, before delivering it to the destination. The default value is 300.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs between 1 and 128, before delivering it to the destination. The default value is 5.
* `catalog_arn` - (Required) Glue catalog ARN identifier of the destination Apache Iceberg Tables. You must specify the ARN in the format `arn:aws:glue:region:account-id:catalog`.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `destination_table_configuration` - (Optional) Destination table configurations which Firehose uses to deliver data to Apache Iceberg Tables. Firehose will write data with insert if table specific configuration is not provided. See [`destination_table_configuration` block](#destination_table_configuration-block) below for details.
* `processing_configuration` - (Optional) Data processing configuration.  See [`processing_configuration` block](#processing_configuration-block) below for details.
* `retry_duration` - (Optional) Period of time, in seconds between 0 to 7200, during which Firehose retries to deliver data to the specified destination.
* `role_arn` - (Required) ARN of the IAM role to be assumed by Firehose for calling Apache Iceberg Tables.
* `s3_backup_mode` - (Optional) Amazon S3 backup mode. Valid values are `FailedDataOnly` and `AllData`. Default value is `FailedDataOnly`.
* `s3_configuration` - (Required) S3 Configuration. See [`s3_configuration` block](#s3_configuration-block) below for details.

### `opensearch_configuration` Block

The `opensearch_configuration` configuration block supports the following arguments:

* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds between 0 to 900, before delivering it to the destination.  The default value is 300s.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs between 1 to 100, before delivering it to the destination.  The default value is 5MB.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `cluster_endpoint` - (Optional) Endpoint to use when communicating with the cluster. Conflicts with `domain_arn`.
* `document_id_options` - (Optional) Method for setting up document ID. See [`document_id_options` block](#document_id_options-block) below for details.
* `domain_arn` - (Optional) ARN of the Amazon ES domain.  The pattern needs to be `arn:.*`.  Conflicts with `cluster_endpoint`.
* `index_name` - (Required) OpenSearch index name.
* `index_rotation_period` - (Optional) OpenSearch index rotation period.  Index rotation appends a timestamp to the IndexName to facilitate expiration of old data.  Valid values are `NoRotation`, `OneHour`, `OneDay`, `OneWeek`, and `OneMonth`.  The default value is `OneDay`.
* `processing_configuration` - (Optional) Data processing configuration. See [`processing_configuration` block](#processing_configuration-block) below for details.
* `retry_duration` - (Optional) After an initial failure to deliver to Amazon OpenSearch, the total amount of time, in seconds between 0 to 7200, during which Firehose re-attempts delivery (including the first attempt).  After this time has elapsed, the failed documents are written to Amazon S3.  The default value is 300s.  There will be no retry if the value is 0.
* `role_arn` - (Required) ARN of the IAM role to be assumed by Firehose for calling the Amazon ES Configuration API and for indexing documents.  The IAM role must have permission for `DescribeDomain`, `DescribeDomains`, and `DescribeDomainConfig`.  The pattern needs to be `arn:.*`.
* `s3_backup_mode` - (Optional) How documents should be delivered to Amazon S3.  Valid values are `FailedDocumentsOnly` and `AllDocuments`.  Default value is `FailedDocumentsOnly`.
* `s3_configuration` - (Required) S3 Configuration. See [`s3_configuration` block](#s3_configuration-block) below for details.
* `type_name` - (Optional) Elasticsearch type name with maximum length of 100 characters. Types are deprecated in OpenSearch_1.1. TypeName must be empty.
* `vpc_config` - (Optional) VPC configuration for the delivery stream to connect to OpenSearch associated with the VPC. See [`vpc_config` block](#vpc_config-block) below for details.

### `opensearchserverless_configuration` Block

The `opensearchserverless_configuration` configuration block supports the following arguments:

* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds between 0 to 900, before delivering it to the destination.  The default value is 300s.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs between 1 to 100, before delivering it to the destination.  The default value is 5MB.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `collection_endpoint` - (Required) Endpoint to use when communicating with the collection in the Serverless offering for Amazon OpenSearch Service.
* `index_name` - (Required) Serverless offering for Amazon OpenSearch Service index name.
* `processing_configuration` - (Optional) Data processing configuration.  See [`processing_configuration` block](#processing_configuration-block) below for details.
* `retry_duration` - (Optional) After an initial failure to deliver to the Serverless offering for Amazon OpenSearch Service, the total amount of time, in seconds between 0 to 7200, during which Kinesis Data Firehose retries delivery (including the first attempt).  After this time has elapsed, the failed documents are written to Amazon S3.  The default value is 300s.  There will be no retry if the value is 0.
* `role_arn` - (Required) ARN of the IAM role to be assumed by Kinesis Data Firehose for calling the Serverless offering for Amazon OpenSearch Service Configuration API and for indexing documents.  The pattern needs to be `arn:.*`.
* `s3_backup_mode` - (Optional) How documents should be delivered to Amazon S3.  Valid values are `FailedDocumentsOnly` and `AllDocuments`.  Default value is `FailedDocumentsOnly`.
* `s3_configuration` - (Required) S3 Configuration. See [`s3_configuration` block](#s3_configuration-block) below for details.
* `vpc_config` - (Optional) VPC configuration for the delivery stream to connect to OpenSearch Serverless associated with the VPC. See [`vpc_config` block](#vpc_config-block) below for details.

### `splunk_configuration` Block

The `splunk_configuration` configuration block supports the following arguments:

* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds between 0 to 60, before delivering it to the destination.  The default value is 60s.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs between 1 to 5, before delivering it to the destination.  The default value is 5MB.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `hec_acknowledgment_timeout` - (Optional) Amount of time, in seconds between 180 and 600, that Kinesis Firehose waits to receive an acknowledgment from Splunk after it sends it data.
* `hec_endpoint` - (Required) HTTP Event Collector (HEC) endpoint to which Kinesis Firehose sends your data.
* `hec_endpoint_type` - (Optional) HEC endpoint type. Valid values are `Raw` or `Event`. The default value is `Raw`.
* `hec_token` - (Optional) GUID that you obtain from your Splunk cluster when you create a new HEC endpoint. This value is required if `secrets_manager_configuration` is not provided.
* `processing_configuration` - (Optional) Data processing configuration.  See [`processing_configuration` block](#processing_configuration-block) below for details.
* `retry_duration` - (Optional) After an initial failure to deliver to Splunk, the total amount of time, in seconds between 0 to 7200, during which Firehose re-attempts delivery (including the first attempt).  After this time has elapsed, the failed documents are written to Amazon S3.  The default value is 300s.  There will be no retry if the value is 0.
* `s3_backup_mode` - (Optional) How documents should be delivered to Amazon S3.  Valid values are `FailedEventsOnly` and `AllEvents`.  Default value is `FailedEventsOnly`.
* `s3_configuration` - (Required) S3 Configuration. See [`s3_configuration` block](#s3_configuration-block) below for details.
* `secrets_manager_configuration` - (Optional) Secrets Manager configuration. See [`secrets_manager_configuration` block](#secrets_manager_configuration-block) below for details. This value is required if `hec_token` is not provided.

### `http_endpoint_configuration` Block

The `http_endpoint_configuration` configuration block supports the following arguments:

* `access_key` - (Optional) Access key required for Kinesis Firehose to authenticate with the HTTP endpoint selected as the destination.
* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds, before delivering it to the destination. The default value is 300 (5 minutes).
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs, before delivering it to the destination. The default value is 5.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `name` - (Optional) HTTP endpoint name.
* `processing_configuration` - (Optional) Data processing configuration.  See [`processing_configuration` block](#processing_configuration-block) below for details.
* `request_configuration` - (Optional) Request configuration.  See [`request_configuration` block](#request_configuration-block) below for details.
* `retry_duration` - (Optional) Total amount of seconds Firehose spends on retries. This duration starts after the initial attempt fails, It does not include the time periods during which Firehose waits for acknowledgment from the specified destination after each attempt. Valid values between `0` and `7200`. Default is `300`.
* `role_arn` - (Required) Kinesis Data Firehose uses this IAM role for all the permissions that the delivery stream needs. The pattern needs to be `arn:.*`.
* `s3_backup_mode` - (Optional) How documents should be delivered to Amazon S3.  Valid values are `FailedDataOnly` and `AllData`.  Default value is `FailedDataOnly`.
* `s3_configuration` - (Required) S3 Configuration. See [`s3_configuration` block](#s3_configuration-block) below for details.
* `secrets_manager_configuration` - (Optional) Secret Manager Configuration. See [`secrets_manager_configuration` block](#secrets_manager_configuration-block) below for details.
* `url` - (Required) HTTP endpoint URL to which Kinesis Firehose sends your data. Refer to the target vendor's documentation for the correct intake URL (for example, [New Relic](https://docs.newrelic.com/docs/infrastructure/amazon-integrations/connect/aws-firehose/) or [Datadog](https://docs.datadoghq.com/integrations/amazon_kinesis_data_firehose/)).

### `snowflake_configuration` Block

The `snowflake_configuration` configuration block supports the following arguments:

* `account_url` - (Required) URL of the Snowflake account. Format: https://[account_identifier].snowflakecomputing.com.
* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds between 0 to 900, before delivering it to the destination.  The default value is 0s.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs between 1 to 128, before delivering it to the destination.  The default value is 1MB.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `content_column_name` - (Optional) Name of the content column.
* `data_loading_option` - (Optional) Data loading option.
* `database` - (Required) Snowflake database name.
* `key_passphrase` - (Optional) Passphrase for the private key.
* `metadata_column_name` - (Optional) Name of the metadata column.
* `private_key` - (Optional) Private key for authentication. This value is required if `secrets_manager_configuration` is not provided.
* `processing_configuration` - (Optional) Processing configuration. See [`processing_configuration` block](#processing_configuration-block) below for details.
* `retry_duration` - (Optional) After an initial failure to deliver to Snowflake, the total amount of time, in seconds between 0 to 7200, during which Firehose re-attempts delivery (including the first attempt).  After this time has elapsed, the failed documents are written to Amazon S3.  The default value is 60s.  There will be no retry if the value is 0.
* `role_arn` - (Required) ARN of the IAM role.
* `s3_backup_mode` - (Optional) S3 backup mode.
* `s3_configuration` - (Required) S3 configuration. See [`s3_configuration` block](#s3_configuration-block) below for details.
* `schema` - (Required) Snowflake schema name.
* `secrets_manager_configuration` - (Optional) Secrets Manager configuration. See [`secrets_manager_configuration` block](#secrets_manager_configuration-block) below for details. This value is required if `user` and `private_key` are not provided.
* `snowflake_role_configuration` - (Optional) Configuration for Snowflake role. See [`snowflake_role_configuration` block](#snowflake_role_configuration-block) below for details.
* `snowflake_vpc_configuration` - (Optional) VPC configuration for Snowflake. See [`snowflake_vpc_configuration` block](#snowflake_vpc_configuration-block) below for details.
* `table` - (Required) Snowflake table name.
* `user` - (Optional) User for authentication. This value is required if `secrets_manager_configuration` is not provided.

### `snowflake_role_configuration` Block

The `snowflake_role_configuration` configuration block supports the following arguments:

* `enabled` - (Optional) Whether the Snowflake role is enabled.
* `snowflake_role` - (Optional) Snowflake role.

### `snowflake_vpc_configuration` Block

The `snowflake_vpc_configuration` configuration block supports the following arguments:

* `private_link_vpce_id` - (Required) VPCE ID for Firehose to privately connect with Snowflake.

### `cloudwatch_logging_options` Block

The `cloudwatch_logging_options` configuration block supports the following arguments:

* `enabled` - (Optional) Enables or disables the logging. Defaults to `false`.
* `log_group_name` - (Optional) CloudWatch group name for logging. This value is required if `enabled` is true.
* `log_stream_name` - (Optional) CloudWatch log stream name for logging. This value is required if `enabled` is true.

### `processing_configuration` Block

The `processing_configuration` configuration block supports the following arguments:

* `enabled` - (Optional) Enables or disables data processing.
* `processors` - (Optional) Data processors as multiple blocks. See [`processors` block](#processors-block) below for details.

### `processors` Block

The `processors` configuration block supports the following arguments:

* `parameters` - (Optional) Processor parameters as multiple blocks. See [`parameters` block](#parameters-block) below for details.
* `type` - (Required) Type of processor. Valid Values: `RecordDeAggregation`, `Lambda`, `MetadataExtraction`, `AppendDelimiterToRecord`, `Decompression`, `CloudWatchLogProcessing`. Validation is done against [AWS SDK constants](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/firehose/types#ProcessorType); so values not explicitly listed may also work.

### `parameters` Block

The `parameters` configuration block supports the following arguments:

* `parameter_name` - (Required) Parameter name. Valid Values: `LambdaArn`, `NumberOfRetries`, `MetadataExtractionQuery`, `JsonParsingEngine`, `RoleArn`, `BufferSizeInMBs`, `BufferIntervalInSeconds`, `SubRecordType`, `Delimiter`, `CompressionFormat`, `DataMessageExtraction`. Validation is done against [AWS SDK constants](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/firehose/types#ProcessorParameterName); so values not explicitly listed may also work.
* `parameter_value` - (Required) Parameter value. Must be between 1 and 512 length (inclusive). When providing a Lambda ARN, you should specify the resource version as well.

~> **NOTE:** Parameters with default values, including `NumberOfRetries`(default: 3), `RoleArn`(default: firehose role ARN), `BufferSizeInMBs`(default: 1), and `BufferIntervalInSeconds`(default: 60), are not stored in terraform state. To prevent perpetual differences, it is therefore recommended to only include parameters with non-default values.

### `request_configuration` Block

The `request_configuration` configuration block supports the following arguments:

* `common_attributes` - (Optional) Metadata sent to the HTTP endpoint destination. See [`common_attributes` block](#common_attributes-block) below for details.
* `content_encoding` - (Optional) Kinesis Data Firehose uses the content encoding to compress the body of a request before sending the request to the destination. Valid values are `NONE` and `GZIP`.  Default value is `NONE`.

### `common_attributes` Block

The `common_attributes` configuration block supports the following arguments:

* `name` - (Required) Name of the HTTP endpoint common attribute.
* `value` - (Required) Value of the HTTP endpoint common attribute.

### `vpc_config` Block

The `vpc_config` configuration block supports the following arguments:

* `role_arn` - (Required) ARN of the IAM role to be assumed by Firehose for calling the Amazon EC2 configuration API and for creating network interfaces. Make sure role has necessary [IAM permissions](https://docs.aws.amazon.com/firehose/latest/dev/controlling-access.html#using-iam-es-vpc).
* `security_group_ids` - (Required) List of security group IDs to associate with Kinesis Firehose.
* `subnet_ids` - (Required) List of subnet IDs to associate with Kinesis Firehose.

### `data_format_conversion_configuration` Block

~> **NOTE:** Once configured, the data format conversion configuration can only be disabled, in which the configuration values will remain, but will not be active. It is not currently possible to completely remove the configuration without recreating the resource.

The `data_format_conversion_configuration` configuration block supports the following arguments:

* `enabled` - (Optional) Defaults to `true`. Set it to `false` if you want to disable format conversion while preserving the configuration details.
* `input_format_configuration` - (Required) Deserializer that you want Kinesis Data Firehose to use to convert the format of your data from JSON. See [`input_format_configuration` block](#input_format_configuration-block) below for details.
* `output_format_configuration` - (Required) Serializer that you want Kinesis Data Firehose to use to convert the format of your data to the Parquet or ORC format. See [`output_format_configuration` block](#output_format_configuration-block) below for details.
* `schema_configuration` - (Required) AWS Glue Data Catalog table that contains the column information. See [`schema_configuration` block](#schema_configuration-block) below for details.

### `s3_configuration` Block

The `s3_configuration` configuration block supports the following arguments:

* `bucket_arn` - (Required) ARN of the S3 bucket.
* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds, before delivering it to the destination. The default value is 300.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs, before delivering it to the destination. The default value is 5. We recommend setting this value greater than the amount of data you typically ingest into the delivery stream in 10 seconds. For example, if you typically ingest data at 1 MB/sec, set this value to 10 MB or higher.
* `cloudwatch_logging_options` - (Optional) CloudWatch Logging Options for the delivery stream. See [`cloudwatch_logging_options` block](#cloudwatch_logging_options-block) below for details.
* `compression_format` - (Optional) Compression format. If no value is specified, the default is `UNCOMPRESSED`. Other supported values are `GZIP`, `ZIP`, `Snappy`, & `HADOOP_SNAPPY`.
* `error_output_prefix` - (Optional) Prefix added to failed records before writing them to S3. Not currently supported for `redshift` destination. This prefix appears immediately following the bucket name. For information about how to specify this prefix, see [Custom Prefixes for Amazon S3 Objects](https://docs.aws.amazon.com/firehose/latest/dev/s3-prefixes.html).
* `kms_key_arn` - (Optional) KMS key ARN the stream will use to encrypt data. If not set, no encryption will be used.
* `prefix` - (Optional) Extra prefix added in front of the "YYYY/MM/DD/HH" time format prefix that is automatically used for delivered S3 files. Note that if the prefix ends with a slash, it appears as a folder in the S3 bucket.
* `role_arn` - (Required) ARN of the AWS credentials.

### `secrets_manager_configuration` Block

The `secrets_manager_configuration` configuration block supports the following arguments:

* `enabled` - (Optional) Enables or disables the Secrets Manager configuration.
* `role_arn` - (Optional) ARN of the role the stream assumes.
* `secret_arn` - (Optional) ARN of the Secrets Manager secret. This value is required if `enabled` is true.

### `input_format_configuration` Block

The `input_format_configuration` configuration block supports the following arguments:

* `deserializer` - (Required) Which deserializer to use. You can choose either the Apache Hive JSON SerDe or the OpenX JSON SerDe. See [`deserializer` block](#deserializer-block) below for details.

### `deserializer` Block

~> **NOTE:** One of the deserializers must be configured. If no nested configuration needs to occur simply declare as `XXX_json_ser_de = []` or `XXX_json_ser_de {}`.

The `deserializer` configuration block supports the following arguments:

* `hive_json_ser_de` - (Optional) Native Hive / HCatalog JsonSerDe. See [`hive_json_ser_de` block](#hive_json_ser_de-block) below for details.
* `open_x_json_ser_de` - (Optional) OpenX SerDe. See [`open_x_json_ser_de` block](#open_x_json_ser_de-block) below for details.

### `hive_json_ser_de` Block

The `hive_json_ser_de` configuration block supports the following arguments:

* `timestamp_formats` - (Optional) List of how you want Kinesis Data Firehose to parse the date and time stamps that may be present in your input data JSON. To specify these format strings, follow the pattern syntax of JodaTime's DateTimeFormat format strings. For more information, see [Class DateTimeFormat](https://www.joda.org/joda-time/apidocs/org/joda/time/format/DateTimeFormat.html). You can also use the special value millis to parse time stamps in epoch milliseconds. If you don't specify a format, Kinesis Data Firehose uses java.sql.Timestamp::valueOf by default.

### `open_x_json_ser_de` Block

The `open_x_json_ser_de` configuration block supports the following arguments:

* `case_insensitive` - (Optional) When set to true, which is the default, Kinesis Data Firehose converts JSON keys to lowercase before deserializing them.
* `column_to_json_key_mappings` - (Optional) Map of column names to JSON keys that aren't identical to the column names. This is useful when the JSON contains keys that are Hive keywords. For example, timestamp is a Hive keyword. If you have a JSON key named timestamp, set this parameter to `{ ts = "timestamp" }` to map this key to a column named ts.
* `convert_dots_in_json_keys_to_underscores` - (Optional) When set to `true`, specifies that the names of the keys include dots and that you want Kinesis Data Firehose to replace them with underscores. This is useful because Apache Hive does not allow dots in column names. For example, if the JSON contains a key whose name is "a.b", you can define the column name to be "a_b" when using this option. Defaults to `false`.

### `output_format_configuration` Block

The `output_format_configuration` configuration block supports the following arguments:

* `serializer` - (Required) Which serializer to use. You can choose either the ORC SerDe or the Parquet SerDe. See [`serializer` block](#serializer-block) below for details.

#### `serializer` Block

~> **NOTE:** One of the serializers must be configured. If no nested configuration needs to occur simply declare as `XXX_ser_de = []` or `XXX_ser_de {}`.

The `serializer` configuration block supports the following arguments:

* `orc_ser_de` - (Optional) Converts data to the ORC format before storing it in Amazon S3. For more information, see [Apache ORC](https://orc.apache.org/docs/). See [`orc_ser_de` block](#orc_ser_de-block) below for details.
* `parquet_ser_de` - (Optional) Converts data to the Parquet format before storing it in Amazon S3. For more information, see [Apache Parquet](https://parquet.apache.org/docs/). See [`parquet_ser_de` block](#parquet_ser_de-block) below for details.

#### `orc_ser_de` Block

The `orc_ser_de` configuration block supports the following arguments:

* `block_size_bytes` - (Optional) Hadoop Distributed File System (HDFS) block size. This is useful if you intend to copy the data from Amazon S3 to HDFS before querying. The default is 256 MiB and the minimum is 64 MiB. Kinesis Data Firehose uses this value for padding calculations.
* `bloom_filter_columns` - (Optional) List of column names for which you want Kinesis Data Firehose to create bloom filters.
* `bloom_filter_false_positive_probability` - (Optional) Bloom filter false positive probability (FPP). The lower the FPP, the bigger the Bloom filter. The default value is `0.05`, the minimum is `0`, and the maximum is `1`.
* `compression` - (Optional) Compression code to use over data blocks. The default is `SNAPPY`.
* `dictionary_key_threshold` - (Optional) Float that represents the fraction of the total number of non-null rows. To turn off dictionary encoding, set this fraction to a number that is less than the number of distinct keys in a dictionary. To always use dictionary encoding, set this threshold to `1`.
* `enable_padding` - (Optional) Set this to `true` to indicate that you want stripes to be padded to the HDFS block boundaries. This is useful if you intend to copy the data from Amazon S3 to HDFS before querying. The default is `false`.
* `format_version` - (Optional) Version of the file to write. The possible values are `V0_11` and `V0_12`. The default is `V0_12`.
* `padding_tolerance` - (Optional) Float between 0 and 1 that defines the tolerance for block padding as a decimal fraction of stripe size. The default value is `0.05`, which means 5 percent of stripe size. For the default values of 64 MiB ORC stripes and 256 MiB HDFS blocks, the default block padding tolerance of 5 percent reserves a maximum of 3.2 MiB for padding within the 256 MiB block. In such a case, if the available size within the block is more than 3.2 MiB, a new, smaller stripe is inserted to fit within that space. This ensures that no stripe crosses block boundaries and causes remote reads within a node-local task. Kinesis Data Firehose ignores this parameter when `enable_padding` is `false`.
* `row_index_stride` - (Optional) Number of rows between index entries. The default is `10000` and the minimum is `1000`.
* `stripe_size_bytes` - (Optional) Number of bytes in each stripe. The default is 64 MiB and the minimum is 8 MiB.

### `parquet_ser_de` Block

The `parquet_ser_de` configuration block supports the following arguments:

* `block_size_bytes` - (Optional) Hadoop Distributed File System (HDFS) block size. This is useful if you intend to copy the data from Amazon S3 to HDFS before querying. The default is 256 MiB and the minimum is 64 MiB. Kinesis Data Firehose uses this value for padding calculations.
* `compression` - (Optional) Compression code to use over data blocks. The possible values are `UNCOMPRESSED`, `SNAPPY`, and `GZIP`, with the default being `SNAPPY`. Use `SNAPPY` for higher decompression speed. Use `GZIP` if the compression ratio is more important than speed.
* `enable_dictionary_compression` - (Optional) Whether to enable dictionary compression.
* `max_padding_bytes` - (Optional) Maximum amount of padding to apply. This is useful if you intend to copy the data from Amazon S3 to HDFS before querying. The default is `0`.
* `page_size_bytes` - (Optional) Parquet page size. Column chunks are divided into pages. A page is conceptually an indivisible unit (in terms of compression and encoding). The minimum value is 64 KiB and the default is 1 MiB.
* `writer_version` - (Optional) Version of row format to output. The possible values are `V1` and `V2`. The default is `V1`.

### `schema_configuration` Block

The `schema_configuration` configuration block supports the following arguments:

* `catalog_id` - (Optional) ID of the AWS Glue Data Catalog. If you don't supply this, the AWS account ID is used by default.
* `database_name` - (Required) Name of the AWS Glue database that contains the schema for the output data.
* `region` - (Optional) If you don't specify an AWS Region, the default is the current region.
* `role_arn` - (Required) Role that Kinesis Data Firehose can use to access AWS Glue. This role must be in the same account you use for Kinesis Data Firehose. Cross-account roles aren't allowed.
* `table_name` - (Required) AWS Glue table that contains the column information that constitutes your data schema.
* `version_id` - (Optional) Table version for the output data schema. Defaults to `LATEST`.

### `destination_table_configuration` Block

The `destination_table_configuration` configuration block supports the following arguments:

* `database_name` - (Required) Name of the Apache Iceberg database.
* `s3_error_output_prefix` - (Optional) Table specific S3 error output prefix. All the errors that occurred while delivering to this table will be prefixed with this value in S3 destination.
* `table_name` - (Required) Name of the Apache Iceberg Table.
* `unique_keys` - (Optional) List of unique keys for a given Apache Iceberg table. Firehose will use these for running Create, Update, or Delete operations on the given Iceberg table.

### `dynamic_partitioning_configuration` Block

The `dynamic_partitioning_configuration` configuration block supports the following arguments:

* `enabled` - (Optional) Enables or disables dynamic partitioning. Defaults to `false`.
* `retry_duration` - (Optional) Total amount of seconds Firehose spends on retries. Valid values between 0 and 7200. Default is 300.

~> **NOTE:** You can enable dynamic partitioning only when you create a new delivery stream. Once you enable dynamic partitioning on a delivery stream, it cannot be disabled on this delivery stream. Therefore, Terraform will recreate the resource whenever dynamic partitioning is enabled or disabled.

### `document_id_options` Block

The `document_id_options` configuration block supports the following arguments:

* `default_document_id_format` - (Required) Method for setting up document ID. Valid values: `FIREHOSE_DEFAULT`, `NO_DOCUMENT_ID`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN specifying the Stream
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `vpc_config` Block

The `vpc_config` configuration block exports the following attributes:

* `vpc_id` - ID of the VPC associated with the delivery stream.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `10m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_kinesis_firehose_delivery_stream.example
  identity = {
    arn = "arn:aws:firehose:us-east-1:123456789012:deliverystream/example-delivery-stream"
  }
}

resource "aws_kinesis_firehose_delivery_stream" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the delivery stream.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Kinesis Firehose Delivery Streams using `arn`. For example:

```terraform
import {
  to = aws_kinesis_firehose_delivery_stream.example
  id = "arn:aws:firehose:us-east-1:123456789012:deliverystream/example-delivery-stream"
}
```

Using `terraform import`, import Kinesis Firehose Delivery Streams using `arn`. For example:

```console
% terraform import aws_kinesis_firehose_delivery_stream.example arn:aws:firehose:us-east-1:123456789012:deliverystream/example-delivery-stream
```

Note: Import does not work for stream destination `s3`. Consider using `extended_s3` since `s3` destination is deprecated.
