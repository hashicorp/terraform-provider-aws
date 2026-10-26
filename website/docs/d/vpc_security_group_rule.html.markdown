---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_rule"
description: |-
    Provides details about a specific security group rule
---

# Data Source: aws_vpc_security_group_rule

`aws_vpc_security_group_rule` provides details about a specific security group rule.

The arguments of this data source act as filters for querying the available security group rules. The given filters must match exactly one security group rule whose data will be exported as attributes.

## Example Usage

```terraform
data "aws_vpc_security_group_rule" "example" {
  security_group_rule_id = var.security_group_rule_id
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `security_group_rule_id` - (Optional) ID of the security group rule to select.

### `filter` Block

* `name` - (Required) Name of the filter field. Valid values can be found in the EC2 [`DescribeSecurityGroupRules`](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroupRules.html) API Reference.
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the security group rule.
* `cidr_ipv4` - Destination IPv4 CIDR range.
* `cidr_ipv6` - Destination IPv6 CIDR range.
* `description` - Security group rule description.
* `from_port` - Start of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 type.
* `ip_protocol` - IP protocol name or number. Use `-1` to specify all protocols.
* `is_egress` - Whether the security group rule is an outbound rule.
* `prefix_list_id` - ID of the destination prefix list.
* `referenced_security_group_id` - Destination security group that is referenced in the rule.
* `security_group_id` - ID of the security group.
* `tags` - Map of tags assigned to the resource.
* `to_port` - End of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 code.
