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

The following arguments are required:

- `target_group_identifier` - (Required) The ID or Amazon Resource Name (ARN) of the target group.
- `target` - (Required) The target.

`target` supports the following:

- `id` - (Required) The ID of the target. If the target type of the target group is INSTANCE, this is an instance ID. If the target type is IP , this is an IP address. If the target type is LAMBDA, this is the ARN of the Lambda function. If the target type is ALB, this is the ARN of the Application Load Balancer.
- `port` - (Optional) The port on which the target is listening. For HTTP, the default is 80. For HTTPS, the default is 443.

## Attribute Reference

This resource exports no additional attributes.
