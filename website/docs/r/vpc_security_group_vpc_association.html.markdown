---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_vpc_association"
description: |-
  Terraform resource for managing Security Group VPC Associations.
---

# Resource: aws_vpc_security_group_vpc_association

Terraform resource for managing Security Group VPC Associations.

## Example Usage

```terraform
resource "aws_vpc_security_group_vpc_association" "example" {
  security_group_id = "sg-05f1f54ab49bb39a3"
  vpc_id            = "vpc-01df9d105095412ba"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `security_group_id` - (Required) The ID of the security group.
* `vpc_id` - (Required) The ID of the VPC to make the association with.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `state` - State of the VPC association. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SecurityGroupVpcAssociation.html) for possible values.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Security Group VPC Association using the `security_group_id` and `vpc_id` arguments, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_vpc_security_group_vpc_association.example
  id = "sg-12345,vpc-67890"
}
```

Using `terraform import`, import a Security Group VPC Association using the `security_group_id` and `vpc_id` arguments, separated by a comma (`,`). For example:

```console
% terraform import aws_vpc_security_group_vpc_association.example sg-12345,vpc-67890
```
