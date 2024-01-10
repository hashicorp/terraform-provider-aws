---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_attachment"
description: |-
  Terraform resource for managing an AWS Auto Scaling Attachment.
---

# Resource: aws_autoscaling_attachment

Attaches a load balancer to an Auto Scaling group.

~> **NOTE on Auto Scaling Groups, Attachments and Traffic Source Attachments:** Terraform provides standalone Attachment (for attaching Classic Load Balancers and Application Load Balancer, Gateway Load Balancer, or Network Load Balancer target groups) and [Traffic Source Attachment](autoscaling_traffic_source_attachment.html) (for attaching Load Balancers and VPC Lattice target groups) resources and an [Auto Scaling Group](autoscaling_group.html) resource with `load_balancers`, `target_group_arns` and `traffic_source` attributes. Do not use the same traffic source in more than one of these resources. Doing so will cause a conflict of attachments. A [`lifecycle` configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html) can be used to suppress differences if necessary.

## Example Usage

```terraform
# Create a new load balancer attachment
resource "aws_autoscaling_attachment" "example" {
  autoscaling_group_name = aws_autoscaling_group.example.id
  elb                    = aws_elb.example.id
}
```

```terraform
# Create a new ALB Target Group attachment
resource "aws_autoscaling_attachment" "example" {
  autoscaling_group_name = aws_autoscaling_group.example.id
  lb_target_group_arn    = aws_lb_target_group.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `autoscaling_group_name` - (Required) Name of ASG to associate with the ELB.
* `elb` - (Optional) Name of the ELB.
* `lb_target_group_arn` - (Optional) ARN of a load balancer target group.

## Attribute Reference

This resource exports no additional attributes.
