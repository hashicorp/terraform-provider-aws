---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_traffic_attachment"
description: |-
  Terraform resource for managing an AWS Auto Scaling Autoscaling Traffic Attachment.
---

# Resource: aws_autoscaling_traffic_attachment

Attaches traffic sources to the specified Auto Scaling group.

## Example Usage

### Basic Usage

```terraform
resource "aws_autoscaling_traffic_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id

  traffic_source {
    identifier = aws_lb_target_group.test[0].arn
    type       = "elbv2"
  }
}
```

## Argument Reference

The following arguments are required:

- `autoscaling_group_name` - (Required) The name of the Auto Scaling group.

- `traffic_source` - (Required) The unique identifiers of a traffic sources.

`traffic_source` supports the following:

- `identifier` - (Required) Identifies the traffic source.For Application Load Balancers, Gateway Load Balancers, Network Load Balancers, and VPC Lattice, this will be the Amazon Resource Name (ARN) for a target group in this account and Region. For Classic Load Balancers, this will be the name of the Classic Load Balancer in this account and Region.
- `type` - (Required) Provides additional context for the value of Identifier.
  The following lists the valid values:
  elb if Identifier is the name of a Classic Load Balancer.
  elbv2 if Identifier is the ARN of an Application Load Balancer, Gateway Load Balancer, or Network Load Balancer target group.
  vpc-lattice if Identifier is the ARN of a VPC Lattice target group.

## Attributes Reference

No additional attributes are exported.
