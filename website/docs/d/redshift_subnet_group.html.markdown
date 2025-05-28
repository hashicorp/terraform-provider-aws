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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the cluster subnet group for which information is requested.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Redshift Subnet Group name.
* `description` - Description of the Redshift Subnet group.
* `id` - Redshift Subnet group Name.
* `subnet_ids` - An array of VPC subnet IDs.
* `tags` - Tags associated to the Subnet Group
