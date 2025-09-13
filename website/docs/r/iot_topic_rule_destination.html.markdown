---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_topic_rule_destination"
description: |-
    Creates and manages an AWS IoT topic rule destination
---

# Resource: aws_iot_topic_rule_destination

## Example Usage

```terraform
resource "aws_iot_topic_rule_destination" "example" {
  vpc_configuration {
    role_arn        = aws_iam_role.example.arn
    security_groups = [aws_security_group.example.id]
    subnet_ids      = aws_subnet.example[*].id
    vpc_id          = aws_vpc.example.id
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `enabled` - (Optional) Whether or not to enable the destination. Default: `true`.
* `vpc_configuration` - (Required) Configuration of the virtual private cloud (VPC) connection. For more info, see the [AWS documentation](https://docs.aws.amazon.com/iot/latest/developerguide/vpc-rule-action.html).

The `vpc_configuration` object takes the following arguments:

* `role_arn` - (Required) The ARN of a role that has permission to create and attach to elastic network interfaces (ENIs).
* `security_groups` - (Optional) The security groups of the VPC destination.
* `subnet_ids` - (Required) The subnet IDs of the VPC destination.
* `vpc_id` - (Required) The ID of the VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the topic rule destination

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IoT topic rule destinations using the `arn`. For example:

```terraform
import {
  to = aws_iot_topic_rule_destination.example
  id = "arn:aws:iot:us-west-2:123456789012:ruledestination/vpc/2ce781c8-68a6-4c52-9c62-63fe489ecc60"
}
```

Using `terraform import`, import IoT topic rule destinations using the `arn`. For example:

```console
% terraform import aws_iot_topic_rule_destination.example arn:aws:iot:us-west-2:123456789012:ruledestination/vpc/2ce781c8-68a6-4c52-9c62-63fe489ecc60
```
