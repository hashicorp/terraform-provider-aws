---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_parameter_group"
description: |-
  Information about a database parameter group.
---

# Data Source: aws_db_parameter_group

Information about a database parameter group.

## Example Usage

```terraform
data "aws_db_parameter_group" "test" {
  name = "default.postgres15"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) DB parameter group name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the parameter group.
* `family` - Family of the parameter group.
* `description` - Description of the parameter group.
