---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_tenant_database"
description: |-
  Lists RDS (Relational Database) Tenant Database resources.
---

# List Resource: aws_rds_tenant_database

Lists RDS (Relational Database) Tenant Database resources.

## Example Usage

### Basic Usage

```terraform
list "aws_rds_tenant_database" "example" {
  provider = aws
}
```

### Tenant Databases in a DB Instance

```terraform
list "aws_rds_tenant_database" "example" {
  provider = aws

  config {
    db_instance_identifier = "example"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `db_instance_identifier` - (Optional) Identifier of the DB instance whose tenant databases to list. Defaults to all DB instances.
* `region` - (Optional) Region to query. Defaults to provider region.
