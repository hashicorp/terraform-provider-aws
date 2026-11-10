---
subcategory: "MemoryDB"
layout: "aws"
page_title: "AWS: aws_memorydb_parameter_group"
description: |-
  Provides information about a MemoryDB Parameter Group.
---

# Data Source: aws_memorydb_parameter_group

Provides information about a MemoryDB Parameter Group.

## Example Usage

```terraform
data "aws_memorydb_parameter_group" "example" {
  name = "my-parameter-group"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the parameter group.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the parameter group.
* `description` - Description of the parameter group.
* `family` - Engine version that the parameter group can be used with.
* `id` - Name of the parameter group.
* `parameter` - Set of user-defined MemoryDB parameters applied by the parameter group.
    * `name` - Name of the parameter.
    * `value` - Value of the parameter.
* `tags` - Map of tags assigned to the parameter group.
