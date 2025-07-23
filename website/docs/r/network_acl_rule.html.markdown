---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_acl_rule"
description: |-
  Provides an network ACL Rule resource.
---

# Resource: aws_network_acl_rule

Creates an entry (a rule) in a network ACL with the specified rule number.

~> **NOTE on Network ACLs and Network ACL Rules:** Terraform currently
provides both a standalone Network ACL Rule resource and a [Network ACL](network_acl.html) resource with rules
defined in-line. At this time you cannot use a Network ACL with in-line rules
in conjunction with any Network ACL Rule resources. Doing so will cause
a conflict of rule settings and will overwrite rules.

## Example Usage

```terraform
resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_network_acl_rule" "bar" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = aws_vpc.foo.cidr_block
  from_port      = 22
  to_port        = 22
}
```

~> **Note:** One of either `cidr_block` or `ipv6_cidr_block` is required.

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `network_acl_id` - (Required) The ID of the network ACL.
* `rule_number` - (Required) The rule number for the entry (for example, 100). ACL entries are processed in ascending order by rule number.
* `egress` - (Optional, bool) Indicates whether this is an egress rule (rule is applied to traffic leaving the subnet). Default `false`.
* `protocol` - (Required) The protocol. A value of -1 means all protocols.
* `rule_action` - (Required) Indicates whether to allow or deny the traffic that matches the rule. Accepted values: `allow` | `deny`
* `cidr_block` - (Optional) The network range to allow or deny, in CIDR notation (for example 172.16.0.0/24 ).
* `ipv6_cidr_block` - (Optional) The IPv6 CIDR block to allow or deny.
* `from_port` - (Optional) The from port to match.
* `to_port` - (Optional) The to port to match.
* `icmp_type` - (Optional) ICMP protocol: The ICMP type. Required if specifying ICMP for the protocolE.g., -1
* `icmp_code` - (Optional) ICMP protocol: The ICMP code. Required if specifying ICMP for the protocolE.g., -1

~> **NOTE:** If the value of `protocol` is `-1` or `all`, the `from_port` and `to_port` values will be ignored and the rule will apply to all ports.

~> **NOTE:** If the value of `icmp_type` is `-1` (which results in a wildcard ICMP type), the `icmp_code` must also be set to `-1` (wildcard ICMP code).

~> Note: For more information on ICMP types and codes, see here: https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the network ACL Rule

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import individual rules using `NETWORK_ACL_ID:RULE_NUMBER:PROTOCOL:EGRESS`, where `PROTOCOL` can be a decimal (such as "6") or string (such as "tcp") value. For example:

**NOTE:** If importing a rule previously provisioned by Terraform, the `PROTOCOL` must be the input value used at creation time. For more information on protocol numbers and keywords, see here: https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml.

Using the procotol's string value:

```terraform
import {
  to = aws_network_acl_rule.my_rule
  id = "acl-7aaabd18:100:tcp:false"
}
```

Using the procotol's decimal value:

```terraform
import {
  to = aws_network_acl_rule.my_rule
  id = "acl-7aaabd18:100:6:false"
}
```

**Using `terraform import` to import** individual rules using `NETWORK_ACL_ID:RULE_NUMBER:PROTOCOL:EGRESS`, where `PROTOCOL` can be a decimal (such as "6") or string (such as "tcp") value. For example:

Using the procotol's string value:

```console
% terraform import aws_network_acl_rule.my_rule acl-7aaabd18:100:tcp:false
```

Using the procotol's decimal value:

```console
% terraform import aws_network_acl_rule.my_rule acl-7aaabd18:100:6:false
```
