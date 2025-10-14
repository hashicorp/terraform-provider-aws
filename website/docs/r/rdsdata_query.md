---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rdsdata_query"
description: |-
  Executes a SQL query against an RDS cluster using the RDS Data API.
---

# Resource: aws_rdsdata_query

Executes a SQL query against an RDS cluster using the RDS Data API. The query is executed once during resource creation, and any changes to the query or parameters will trigger a replacement.

~> **Note:** For queries that need to be executed multiple times or for retrieving data (SELECT queries), consider using the [`aws_rdsdata_query` data source](/docs/providers/aws/d/rdsdata_query.html) instead. Use this resource for one-time operations like DDL statements, INSERT, UPDATE, or DELETE operations.

## Example Usage

### Basic Usage

```terraform
resource "aws_rdsdata_query" "example" {
  resource_arn = aws_rds_cluster.example.arn
  secret_arn   = aws_secretsmanager_secret.example.arn
  sql          = "SELECT * FROM users WHERE active = true"
  database     = "mydb"
}
```

### With Parameters

```terraform
resource "aws_rdsdata_query" "example" {
  resource_arn = aws_rds_cluster.example.arn
  secret_arn   = aws_secretsmanager_secret.example.arn
  sql          = "INSERT INTO users (name, email) VALUES (:name, :email)"
  database     = "mydb"

  parameters {
    name  = "name"
    value = "John Doe"
  }

  parameters {
    name  = "email"
    value = "john@example.com"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the RDS cluster.
* `secret_arn` - (Required) The ARN of the secret that enables access to the DB cluster.
* `sql` - (Required) The SQL statement to execute.
* `database` - (Optional) The name of the database.
* `parameters` - (Optional) Parameters for the SQL statement. See [parameters](#parameters) below.
* `region` - (Optional) The AWS region.

### parameters

* `name` - (Required) The name of the parameter.
* `value` - (Required) The value of the parameter.
* `type_hint` - (Optional) A hint that specifies the correct object type for the parameter value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The resource identifier.
* `records` - The records returned by the SQL statement in JSON format.
* `number_of_records_updated` - The number of records updated by the statement.

## Import

You cannot import this resource.
