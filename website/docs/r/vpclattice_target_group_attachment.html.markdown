---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_target_group_attachment"
description: |-
  Provides the ability to register a target with an AWS VPC Lattice Target Group.
---

# Resource: aws_vpclattice_target_group_attachment

Provides the ability to register a target with an AWS VPC Lattice Target Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_target_group_attachment" "example" {
  target_group_identifier = aws_vpclattice_target_group.example.id

  target {
    id   = aws_lb.example.arn
    port = 80
  }
}
```

## Argument Reference

This resource supports the following arguments:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `target_group_identifier` - (Required) The ID or Amazon Resource Name (ARN) of the target group.
- `target` - (Required) The target.

`target` supports the following:

- `id` - (Required) The ID of the target. If the target type of the target group is INSTANCE, this is an instance ID. If the target type is IP , this is an IP address. If the target type is LAMBDA, this is the ARN of the Lambda function. If the target type is ALB, this is the ARN of the Application Load Balancer.
- `port` - (Optional) This port is used for routing traffic to the target, and defaults to the target group port. However, you can override the default and specify a custom port.

## Attribute Reference

This resource exports no additional attributes.
