---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_engine_versions"
description: |-
  List of RDS engine versions.
---

# Data Source: aws_rds_engine_version

List of RDS engine versions.

## Example Usage

### Basic Usage

```terraform
data "aws_rds_engine_versions" "test" {
  filter {
    name   = "engine"
    values = ["aurora-mysql"]
  }
  filter {
    name   = "db-parameter-group-family"
    values = ["aurora-mysql8.0"]
  }
  filter {
    name   = "engine-mode"
    values = ["provisioned"]
  }
}
locals {
  versions       = data.aws_rds_engine_versions.test.versions
  latest_version = element(local.versions, length(local.versions) - 1)
}
data "aws_rds_cluster" "test" {
  engine         = "aurora-mysql"
  engine_version = local.latest_version.engine_version
  # ...
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) One or more name/value pairs to filter off of. There are several valid keys; for a full reference, check out [describe-db-engine-versions in the AWS CLI reference](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/rds/describe-db-engine-versions.html).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `versions` - List of all matching RDS engine versions. Each item contains:
    * `engine`
    * `engine_version`
    * `db_parameter_group_family`
    * `status` - Status of this version, either 'available' or 'deprecated'.
