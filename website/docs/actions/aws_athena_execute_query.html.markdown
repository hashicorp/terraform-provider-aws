---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_execute_query"
description: Executes an Athena query on-demand.
---

# Action aws_athena_execute_query

Executes an Athena query on-demand. This action will initiate an Athena query and wait for it to complete, providing progress updates during execution.

For information about Athena queries, see the [Use Athena SQL guide](https://docs.aws.amazon.com/athena/latest/ug/using-athena-sql.html). For specific information about running Athena queries, see the [StartQueryExecution](https://docs.aws.amazon.com/athena/latest/APIReference/API_StartQueryExecution.html) page in the Athena API Reference.

## Example Usage

### Basic Usage

```terraform
action "aws_athena_execute_query" "query" {
  config {
    query_string = <<EOT
CREATE EXTERNAL TABLE test (
    test INT
    )
LOCATION 's3://${aws_s3_bucket.test.id}/test/'
EOT
    workgroup    = "primary"
    query_execution_context {
      database = aws_glue_catalog_database.database.name
    }
    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.id}/query-results/"
    }
  }
}

resource "aws_glue_catalog_database" "database" {
  name = "test"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "bucket"
}

resource "terraform_data" "trigger" {
  input = "trigger-query"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_athena_execute_query.query]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `query_string` - (Required) The SQL query statements to be executed.

The following arguments are optional:

* `execution_parameters` - (Optional) A list of values for the parameters in a query. The values are applied sequentially to the parameters in the query in the order in which the parameters occur.
* `query_execution_context` - (Optional) The database and data catalog context in which the query execution occurs. See [query_execution_context](#query-execution-context) below.
* `result_configuration` - (Optional) Specifies information about where and how to save the results of the query execution. See [result_configuration](#result-configuration) below.
* `timeout` - (Optional) Timeout in seconds for the backup operation. Defaults to 300 seconds.
* `workgroup` - (Optional) The name of the workgroup in which the query is being started

### Query Execution Context

The `query_execution_context` block supports the following:

* `catalog` - (Optional) The name of the data catalog used in the query execution.
* `database` - (Optional) The name of the database used in the query execution. The database must exist in the catalog.

### Result Configuration

The `result_configuration` block supports the following:

* `expected_bucket_owner` - (Optional) The AWS account ID that you expect to be the owner of the Amazon S3 bucket.
* `output_location` - (Optional) The location in Amazon S3 where your query and calculation results are stored, such as `s3://path/to/query/bucket/`.
* `acl_configuration` - (Optional) Indicates that an Amazon S3 canned ACL should be set to control ownership of stored query results. See [acl_configuration](#acl-configuration) below.
* `encryption_configuration` - (Optional) If query and calculation results are encrypted in Amazon S3, indicates the encryption option used. See [encryption_configuration](#encryption-configuration) below.

### ACL Configuration

The `acl_configuration` block supports the following:

* `s3_acl_option` - (Optional) The Amazon S3 canned ACL that Athena should specify when storing query results.

### Encryption Configuration

The `encryption_configuration` block supports the following:

* `encryption_option` - (Optional) Indicates whether Amazon S3 server-side encryption with Amazon S3-managed keys (SSE_S3), server-side encryption with KMS-managed keys (SSE_KMS), or client-side encryption with KMS-managed keys (CSE_KMS) is used.
* `kms_key` - (Optional) For SSE_KMS and CSE_KMS, this is the KMS key ARN or ID.
