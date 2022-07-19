---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_subnet_group"
description: |-
  Provides details about a specific redshift subnet_group
---

# Data Source: aws_redshift_subnet_group

Provides details about a specific redshift subnet group.

## Example Usage

```terraform
data "aws_redshift_subnet_group" "example" {
  name = aws_redshift_subnet_group.example.name
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the cluster subnet group for which information is requested.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Redshift Subnet Group name.
* `description` - The description of the Redshift Subnet group.
* `id` - The Redshift Subnet group Name.
* `subnet_ids` - An array of VPC subnet IDs.
* `tags` - The tags associated to the Subnet Group
