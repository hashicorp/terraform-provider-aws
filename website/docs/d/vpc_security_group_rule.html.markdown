---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_rule"
description: |-
    Provides details about a specific security group rule
---

# Data Source: aws_vpc_security_group_rule

`aws_vpc_security_group_rule` provides details about a specific security group rule.

## Example Usage

```terraform
data "aws_vpc_security_group_rule" "example" {
  security_group_rule_id = var.security_group_rule_id
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
security group rules. The given filters must match exactly one security group rule
whose data will be exported as attributes.

* `security_group_rule_id` - (Optional) ID of the security group rule to select.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the EC2 [`DescribeSecurityGroupRules`](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroupRules.html) API Reference.
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the security group rule.
* `cidr_ipv4` - The destination IPv4 CIDR range.
* `cidr_ipv6` - The destination IPv6 CIDR range.
* `description` - The security group rule description.
* `from_port` - The start of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 type.
* `is_egress` - Indicates whether the security group rule is an outbound rule.
* `ip_protocol` - The IP protocol name or number. Use `-1` to specify all protocols.
* `prefix_list_id` - The ID of the destination prefix list.
* `referenced_security_group_id` - The destination security group that is referenced in the rule.
* `security_group_id` - The ID of the security group.
* `tags` - A map of tags assigned to the resource.
* `to_port` - (Optional) The end of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 code.
