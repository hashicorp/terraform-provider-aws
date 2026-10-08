---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_proxies"
description: |-
  Terraform data source for managing an AWS RDS (Relational Database) Proxies.
---

# Data Source: aws_db_proxies

Terraform data source for managing an AWS RDS (Relational Database) DB Proxies.

## Example Usage

### Basic Usage

```terraform
data "aws_db_Proxies" "example" {}
```

## Argument Reference

## Attribute Reference

This data source exports the following attribute:

* `names` - Set of names of the RDS DB proxies.
