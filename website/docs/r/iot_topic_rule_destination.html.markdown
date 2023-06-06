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

* `enabled` - (Optional) Whether or not to enable the destination. Default: `true`.
* `vpc_configuration` - (Required) Configuration of the virtual private cloud (VPC) connection. For more info, see the [AWS documentation](https://docs.aws.amazon.com/iot/latest/developerguide/vpc-rule-action.html).

The `vpc_configuration` object takes the following arguments:

* `role_arn` - (Required) The ARN of a role that has permission to create and attach to elastic network interfaces (ENIs).
* `security_groups` - (Optional) The security groups of the VPC destination.
* `subnet_ids` - (Required) The subnet IDs of the VPC destination.
* `vpc_id` - (Required) The ID of the VPC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the topic rule destination

## Import

IoT topic rule destinations can be imported using the `arn`, e.g.,

```
$ terraform import aws_iot_topic_rule_destination.example arn:aws:iot:us-west-2:123456789012:ruledestination/vpc/2ce781c8-68a6-4c52-9c62-63fe489ecc60
```
