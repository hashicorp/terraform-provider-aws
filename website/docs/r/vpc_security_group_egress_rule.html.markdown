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

~> **NOTE on Security Groups and Security Group Rules:** TODO explain why this is now the prefered resource for managing security group rules.

## Example Usage

```terraform
resource "aws_vpc_security_group_egress_rule" "example" {
  security_group_id = aws_security_group.example.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080
}
```

## Argument Reference

The following arguments are supported:

* `cidr_ipv4` - (Optional) The destination IPv4 CIDR range.
* `cidr_ipv6` - (Optional) The destination IPv6 CIDR range.
* `description` - (Optional) The security group rule description.
* `from_port` - (Optional) The start of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 type.
* `ip_protocol` - (Optional) The IP protocol name or number. Use `-1` to specify all protocols.
* `prefix_list_id` - (Optional) The ID of the destination prefix list.
* `referenced_security_group_id` - (Optional) The destination security group that is referenced in the rule.
* `security_group_id` - (Required) The ID of the security group.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `to_port` - (Optional) The end of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 code.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the security group rule.
* `security_group_rule_id` - The ID of the security group rule.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Security group egress rules can be imported using the `security_group_rule_id`, e.g.,

```
$ terraform import aws_vpc_security_group_egress_rule.example sgr-02108b27edd666983
```