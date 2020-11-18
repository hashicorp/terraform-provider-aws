---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_resource_policy"
description: |-
  Provides an AWS Network Firewall Resource Policy resource.
---

# Resource: aws_networkfirewall_resource_policy

Provides an AWS Network Firewall Resource Policy Resource for a rule group or firewall policy.

## Example Usage

### For a Firewall Policy resource

```hcl
resource "aws_networkfirewall_resource_policy" "example" {
  resource_arn = data.aws_iam_user.example.arn
  policy       = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": [
        "network-firewall:ListFirewallPolicies",
        "network-firewall:AssociateFirewallPolicy"
      ],
      "Resource": [
        "${aws_networkfirewall_firewall_policy.example.arn}"
      ]
    }
  ]
}
POLICY
}
```

### For a Rule Group resource

```hcl
resource "aws_networkfirewall_resource_policy" "example" {
  resource_arn = data.aws_iam_user.example.arn
  policy       = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": [
        "network-firewall:ListRuleGroups"
      ],
      "Resource": [
        "${aws_networkfirewall_rule_group.example.arn}"
      ]
    }
  ]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `policy` - (Required) JSON formatted policy document that controls access to the Network Firewall resource.

* `resource_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the account that you want to share rule groups and firewall policies with.

## Attribute Reference

In addition to all arguments above, the following attribute is exported:

* `id` - The Amazon Resource Name (ARN) of the account associated with the resource policy.

## Import

Network Firewall Resource Policies can be imported using the `resource_arn` e.g.

```
$ terraform import aws_networkfirewall_resource_policy.example arn:aws:iam:1234567890:user/example
```
