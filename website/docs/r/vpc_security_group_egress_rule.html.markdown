---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_egress_rule"
description: |-
  Provides a VPC security group egress rule resource.
---

# Resource: aws_vpc_security_group_egress_rule

Manages an outbound (egress) rule for a security group.

When specifying an outbound rule for your security group in a VPC, the configuration must include a destination for the traffic.

~> **NOTE:** Using `aws_vpc_security_group_egress_rule` and [`aws_vpc_security_group_ingress_rule`](vpc_security_group_ingress_rule.html) resources is the current best practice. Avoid using the [`aws_security_group_rule`](security_group_rule.html) resource and the `ingress` and `egress` arguments of the [`aws_security_group`](security_group.html) resource for configuring in-line rules, as they struggle with managing multiple CIDR blocks, and tags and descriptions due to the historical lack of unique IDs.

!> **WARNING:** You should not use the `aws_vpc_security_group_egress_rule` and [`aws_vpc_security_group_ingress_rule`](vpc_security_group_ingress_rule.html) resources in conjunction with the [`aws_security_group`](security_group.html) resource with _in-line rules_ (using the `ingress` and `egress` arguments of `aws_security_group`) or the [`aws_security_group_rule`](security_group_rule.html) resource. Doing so may cause rule conflicts, perpetual differences, and result in rules being overwritten.

## Example Usage

```terraform
resource "aws_vpc_security_group_egress_rule" "example" {
  security_group_id = aws_security_group.example.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 80
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cidr_ipv4` - (Optional) The destination IPv4 CIDR range.
* `cidr_ipv6` - (Optional) The destination IPv6 CIDR range.
* `description` - (Optional) The security group rule description.
* `from_port` - (Optional) The start of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 type.
* `ip_protocol` - (Optional) The IP protocol name or number. Use `-1` to specify all protocols. Note that if `ip_protocol` is set to `-1`, it translates to all protocols, all port ranges, and `from_port` and `to_port` values should not be defined.
* `prefix_list_id` - (Optional) The ID of the destination prefix list.
* `referenced_security_group_id` - (Optional) The destination security group that is referenced in the rule.
* `security_group_id` - (Required) The ID of the security group.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `to_port` - (Optional) The end of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 code.

~> **Note** Although `cidr_ipv4`, `cidr_ipv6`, `prefix_list_id`, and `referenced_security_group_id` are all marked as optional, you *must* provide one of them in order to configure the destination of the traffic. The `from_port` and `to_port` arguments are required unless `ip_protocol` is set to `-1` or `icmpv6`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the security group rule.
* `security_group_rule_id` - The ID of the security group rule.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import security group egress rules using the `security_group_rule_id`. For example:

```terraform
import {
  to = aws_vpc_security_group_egress_rule.example
  id = "sgr-02108b27edd666983"
}
```

Using `terraform import`, import security group egress rules using the `security_group_rule_id`. For example:

```console
% terraform import aws_vpc_security_group_egress_rule.example sgr-02108b27edd666983
```
