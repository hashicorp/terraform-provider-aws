---
subcategory: "RDS Data"
layout: "aws"
page_title: "AWS: aws_rdsdata_query"
description: |-
  Executes SQL queries against RDS clusters using the RDS Data API.
---

# Data Source: aws_rdsdata_query

Executes SQL queries against RDS clusters using the RDS Data API. This data source allows you to run SQL statements and retrieve results in JSON format.

~> **Note:** This data source is ideal for SELECT queries that need to be executed multiple times during Terraform operations. For one-time operations like DDL statements, INSERT, UPDATE, or DELETE operations, consider using the [`aws_rdsdata_query` resource](/docs/providers/aws/r/rdsdata_query.html) instead.

## Example Usage

### Basic Query

```terraform
data "aws_rdsdata_query" "example" {
  resource_arn = aws_rds_cluster.example.arn
  secret_arn   = aws_secretsmanager_secret.example.arn
  sql          = "SELECT * FROM users LIMIT 10"
}
```

### Query with Parameters

```terraform
data "aws_rdsdata_query" "example" {
  resource_arn = aws_rds_cluster.example.arn
  secret_arn   = aws_secretsmanager_secret.example.arn
  sql          = "SELECT * FROM users WHERE status = :status"
  database     = "myapp"

  parameters {
    name  = "status"
    value = "active"
  }
}
```

### Query with Multiple Parameters

```terraform
data "aws_rdsdata_query" "example" {
  resource_arn = aws_rds_cluster.example.arn
  secret_arn   = aws_secretsmanager_secret.example.arn
  sql          = "SELECT * FROM orders WHERE user_id = :user_id AND created_at > :date"

  parameters {
    name  = "user_id"
    value = "123"
  }

  parameters {
    name  = "date"
    value = "2023-01-01"
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the Aurora Serverless DB cluster.
* `secret_arn` - (Required) The ARN of the secret that enables access to the DB cluster. The secret must contain the database credentials.
* `sql` - (Required) The SQL statement to execute.
* `database` - (Optional) The name of the database to execute the statement against.
* `parameters` - (Optional) Parameters for the SQL statement. See [Parameters](#parameters) below.
* `region` - (Optional) The AWS region where the RDS cluster is located. If not specified, the provider region is used.

### Parameters

The `parameters` block supports the following:

* `name` - (Required) The name of the parameter.
* `value` - (Required) The value of the parameter as a string.
* `type_hint` - (Optional) A hint that specifies the correct object type for the parameter value. Valid values: `DATE`, `DECIMAL`, `JSON`, `TIME`, `TIMESTAMP`, `UUID`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `records` - The records returned by the SQL statement in JSON format.
* `number_of_records_updated` - The number of records updated by the request (for DML statements).

## Notes

* This data source requires the Aurora Serverless cluster to have the Data API enabled.
* The secret must be created in AWS Secrets Manager and contain the database credentials in the correct format.
* Results are returned in JSON format when using `SELECT` statements.
* For non-SELECT statements (INSERT, UPDATE, DELETE), the `number_of_records_updated` attribute will contain the count of affected rows.
* The Data API has a 1MB limit for response data. Large result sets may be truncated.
