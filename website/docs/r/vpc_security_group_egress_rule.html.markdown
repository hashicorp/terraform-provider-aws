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

~> **NOTE on Security Groups and Security Group Rules:** Terraform currently provides a [Security Group resource](security_group.html) with `ingress` and `egress` rules defined in-line and a [Security Group Rule resource](security_group_rule.html) which manages one or more `ingress` or
`egress` rules. Both of these resource were added before AWS assigned a [security group rule unique ID](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/security-group-rules.html), and they do not work well in all scenarios using the`description` and `tags` attributes, which rely on the unique ID.
The `aws_vpc_security_group_egress_rule` resource has been added to address these limitations and should be used for all new security group rules.
You should not use the `aws_vpc_security_group_egress_rule` resource in conjunction with an `aws_security_group` resource with in-line rules or with `aws_security_group_rule` resources defined for the same Security Group, as rule conflicts may occur and rules will be overwritten.

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

~> **Note** Although `cidr_ipv4`, `cidr_ipv6`, `prefix_list_id`, and `referenced_security_group_id` are all marked as optional, you *must* provide one of them in order to configure the destination of the traffic. The `from_port` and `to_port` arguments are required unless `ip_protocol` is set to `-1` or `icmpv6`.

This resource supports the following arguments:

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
