---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_security_group"
description: |-
    Provides details about a specific Security Group
---

# Data Source: aws_security_group

`aws_security_group` provides details about a specific Security Group.

This resource can prove useful when a module accepts a Security Group id as
an input variable and needs to, for example, determine the id of the
VPC that the security group belongs to.

## Example Usage

The following example shows how one might accept a Security Group id as a variable
and use this data source to obtain the data necessary to create a subnet.

```terraform
variable "security_group_id" {}

data "aws_security_group" "selected" {
  id = var.security_group_id
}

resource "aws_subnet" "subnet" {
  vpc_id     = data.aws_security_group.selected.vpc_id
  cidr_block = "10.0.1.0/24"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Custom filter block as described below.
* `id` - (Optional) Id of the specific security group to retrieve.
* `name` - (Optional) Name that the desired security group must have.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired security group.
* `vpc_id` - (Optional) Id of the VPC that the desired security group belongs to.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroups.html).
* `values` - (Required) Set of values that are accepted for the given field.
  A Security Group will be selected if any one of the given values matches.

## Attribute Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected Security Group.

The following fields are also exported:

* `description` - Description of the security group.
* `arn` - Computed ARN of the security group.

~> **Note:** The [default security group for a VPC](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_SecurityGroups.html#DefaultSecurityGroup) has the name `default`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
