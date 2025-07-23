---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_arn"
description: |-
    Parses an ARN into its constituent parts.
---

# Data Source: aws_arn

Parses an ARN into its constituent parts.

## Example Usage

```terraform
data "aws_arn" "db_instance" {
  arn = "arn:aws:rds:eu-west-1:123456789012:db:mysql-db"
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN to parse.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `partition` - Partition that the resource is in.
* `service` - The [service namespace](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#genref-aws-service-namespaces) that identifies the AWS product.
* `region` - Region the resource resides in.
Note that the ARNs for some resources do not include a Region, so this component might be omitted.
* `account` - The [ID](https://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html) of the AWS account that owns the resource, without the hyphens.
* `resource` - Content of this part of the ARN varies by service.
It often includes an indicator of the type of resource—for example, an IAM user or Amazon RDS database —followed by a slash (/) or a colon (:), followed by the resource name itself.
