---
layout: "aws"
page_title: "AWS: aws_athena_database"
sidebar_current: "docs-aws-resource-athena-database"
description: |-
  Provides an Athena database.
---

# aws_athena_database

Provides an Athena database.

## Example Usage

```hcl
resource "aws_s3_bucket" "hoge" {
  bucket = "hoge"
}

resource "aws_athena_database" "hoge" {
  name = "database_name"
  bucket = "${aws_s3_bucket.hoge.bucket}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the database to create.
* `bucket` - (Required) Name of s3 bucket to save the results of the query execution.
* `force_destroy` - (Optional, Default: false) A boolean that indicates all tables should be deleted from the database so that the database can be destroyed without error. The tables are *not* recoverable.

~> **NOTE:** When Athena queries are executed, result files may be created in the specified bucket. Consider using `force_destroy` on the bucket too in order to avoid any problems when destroying the bucket.  

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The database name
