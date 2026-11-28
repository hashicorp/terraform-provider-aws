---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_instance"
description: |-
  Lists RDS (Relational Database) Cluster Instance resources.
---

# List Resource: aws_rds_cluster_instance

Lists RDS (Relational Database) Cluster Instance resources.

## Example Usage

```terraform
list "aws_rds_cluster_instance" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
