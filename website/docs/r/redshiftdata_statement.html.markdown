---
subcategory: "Redshift Data"
layout: "aws"
page_title: "AWS: aws_redshiftdata_.subnet_group"
description: |-
  Provides a Redshift Data Statement execution resource.
---

# Resource: aws_redshiftdata_statement

Executes a Redshift Data Statement.

## Example Usage

```terraform
resource "aws_redshiftdata_statement" "example" {
  cluster_identifier = aws_redshift_cluster.example.cluster_identifier
  database           = aws_redshift_cluster.example.database_name
  db_user            = aws_redshift_cluster.example.master_username
  sql                = "CREATE GROUP group_name;"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_identifier` - (Required) The cluster identifier.
* `database` - (Required) The name of the database.
* `db_user` - (Optional) The database user name.
* `secret_arn` - (Optional) The name or ARN of the secret that enables access to the database.
* `sql` - (Required) The SQL statement text to run.
* `statement_name` - (Optional) The name of the SQL statement. You can name the SQL statement when you create it to identify the query.
* `with_event` - (Optional) A value that indicates whether to send an event to the Amazon EventBridge event bus after the SQL statement runs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Redshift Data Statement ID.

## Import

Redshift Data Statements can be imported using the `id`, e.g.,

```
$ terraform import aws_redshiftdata_statement.example example
```
