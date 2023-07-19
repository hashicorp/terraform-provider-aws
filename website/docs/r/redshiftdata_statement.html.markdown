---
subcategory: "Redshift Data"
layout: "aws"
page_title: "AWS: aws_redshiftdata_statement"
description: |-
  Provides a Redshift Data Statement execution resource.
---

# Resource: aws_redshiftdata_statement

Executes a Redshift Data Statement.

## Example Usage

### cluster_identifier

```terraform
resource "aws_redshiftdata_statement" "example" {
  cluster_identifier = aws_redshift_cluster.example.cluster_identifier
  database           = aws_redshift_cluster.example.database_name
  db_user            = aws_redshift_cluster.example.master_username
  sql                = "CREATE GROUP group_name;"
}
```

### workgroup_name

```terraform
resource "aws_redshiftdata_statement" "example" {
  workgroup_name = aws_redshiftserverless_workgroup.example.workgroup_name
  database       = "dev"
  sql            = "CREATE GROUP group_name;"
}
```

## Argument Reference

The following arguments are required:

* `database` - (Required) The name of the database.
* `sql` - (Required) The SQL statement text to run.

The following arguments are optional:

* `cluster_identifier` - (Optional) The cluster identifier. This parameter is required when connecting to a cluster and authenticating using either Secrets Manager or temporary credentials.
* `db_user` - (Optional) The database user name.
* `secret_arn` - (Optional) The name or ARN of the secret that enables access to the database.
* `statement_name` - (Optional) The name of the SQL statement. You can name the SQL statement when you create it to identify the query.
* `with_event` - (Optional) A value that indicates whether to send an event to the Amazon EventBridge event bus after the SQL statement runs.
* `workgroup_name` - (Optional) The serverless workgroup name. This parameter is required when connecting to a serverless workgroup and authenticating using either Secrets Manager or temporary credentials.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Redshift Data Statement ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Data Statements using the `id`. For example:

```terraform
import {
  to = aws_redshiftdata_statement.example
  id = "example"
}
```

Using `terraform import`, import Redshift Data Statements using the `id`. For example:

```console
% terraform import aws_redshiftdata_statement.example example
```
