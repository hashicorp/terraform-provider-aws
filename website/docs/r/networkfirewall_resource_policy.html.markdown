---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_resource_policy"
description: |-
  Provides an AWS Network Firewall Resource Policy resource.
---

# Resource: aws_networkfirewall_resource_policy

Provides an AWS Network Firewall Resource Policy Resource

## Example Usage

```hcl
resource "aws_networkfirewall_resource_policy" "example" {
  resource_arn = aws_networkfirewall_firewall.example.arn
  policy       = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": [
        "ec2:AttachNetworkInterface",
        "ec2:CreateNetworkInterface",
        "ec2:CreateNetworkInterfacePermission",
        "ec2:DeleteNetworkInterface",
        "ec2:DeleteNetworkInterfacePermission",
        "ec2:DescribeInstances",
        "ec2:DescribeNetworkInterfaceAttribute",
        "ec2:DescribeNetworkInterfacePermissions",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcs",
        "ec2:DetachNetworkInterface",
        "ec2:ModifyNetworkInterfaceAttribute"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `policy` - (Required) JSON formatted policy document that controls access to the Network Firewall resource.

* `resource_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the resource associated with the resource policy.

## Attribute Reference

In addition to all arguments above, the following attribute is exported:

* `id` - The Amazon Resource Name (ARN) of the resource associated with the resource policy.

## Import

Network Firewall Resource Policies can be imported using the `resource_arn` e.g.

```
$ terraform import aws_networkfirewall_resource_policy.example arn:aws:network-firewall:us-west-1:123456789012:firewall/example
```
