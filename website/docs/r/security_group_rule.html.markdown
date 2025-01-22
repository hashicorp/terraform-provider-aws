---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_security_group_rule"
description: |-
  Provides an security group rule resource.
---

# Resource: aws_security_group_rule

Provides a security group rule resource. Represents a single `ingress` or `egress` group rule, which can be added to external Security Groups.

~> **NOTE:** Avoid using the `aws_security_group_rule` resource, as it struggles with managing multiple CIDR blocks, and, due to the historical lack of unique IDs, tags and descriptions. To avoid these problems, use the current best practice of the [`aws_vpc_security_group_egress_rule`](vpc_security_group_egress_rule.html) and [`aws_vpc_security_group_ingress_rule`](vpc_security_group_ingress_rule.html) resources with one CIDR block per rule.

!> **WARNING:** You should not use the `aws_security_group_rule` resource in conjunction with [`aws_vpc_security_group_egress_rule`](vpc_security_group_egress_rule.html) and [`aws_vpc_security_group_ingress_rule`](vpc_security_group_ingress_rule.html) resources or with an [`aws_security_group`](security_group.html) resource that has in-line rules. Doing so may cause rule conflicts, perpetual differences, and result in rules being overwritten.

~> **NOTE:** Setting `protocol = "all"` or `protocol = -1` with `from_port` and `to_port` will result in the EC2 API creating a security group rule with all ports open. This API behavior cannot be controlled by Terraform and may generate warnings in the future.

~> **NOTE:** Referencing Security Groups across VPC peering has certain restrictions. More information is available in the [VPC Peering User Guide](https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-security-groups.html).

## Example Usage

Basic usage

```terraform
resource "aws_security_group_rule" "example" {
  type              = "ingress"
  from_port         = 0
  to_port           = 65535
  protocol          = "tcp"
  cidr_blocks       = [aws_vpc.example.cidr_block]
  ipv6_cidr_blocks  = [aws_vpc.example.ipv6_cidr_block]
  security_group_id = "sg-123456"
}
```

### Usage With Prefix List IDs

Prefix Lists are either managed by AWS internally, or created by the customer using a
[Managed Prefix List resource](ec2_managed_prefix_list.html). Prefix Lists provided by
AWS are associated with a prefix list name, or service name, that is linked to a specific region.

Prefix list IDs are exported on VPC Endpoints, so you can use this format:

```terraform
resource "aws_security_group_rule" "allow_all" {
  type              = "egress"
  to_port           = 0
  protocol          = "-1"
  prefix_list_ids   = [aws_vpc_endpoint.my_endpoint.prefix_list_id]
  from_port         = 0
  security_group_id = "sg-123456"
}

# ...
resource "aws_vpc_endpoint" "my_endpoint" {
  # ...
}
```

You can also find a specific Prefix List using the [`aws_prefix_list`](/docs/providers/aws/d/prefix_list.html)
or [`ec2_managed_prefix_list`](/docs/providers/aws/d/ec2_managed_prefix_list.html) data sources:

```terraform
data "aws_region" "current" {}

data "aws_prefix_list" "s3" {
  name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

resource "aws_security_group_rule" "s3_gateway_egress" {
  # S3 Gateway interfaces are implemented at the routing level which means we
  # can avoid the metered billing of a VPC endpoint interface by allowing
  # outbound traffic to the public IP ranges, which will be routed through
  # the Gateway interface:
  # https://docs.aws.amazon.com/AmazonS3/latest/userguide/privatelink-interface-endpoints.html
  description       = "S3 Gateway Egress"
  type              = "egress"
  security_group_id = "sg-123456"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  prefix_list_ids   = [data.aws_prefix_list.s3.id]
}
```

## Argument Reference

The following arguments are required:

* `from_port` - (Required) Start port (or ICMP type number if protocol is "icmp" or "icmpv6").
* `protocol` - (Required) Protocol. If not icmp, icmpv6, tcp, udp, or all use the [protocol number](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml)
* `security_group_id` - (Required) Security group to apply this rule to.
* `to_port` - (Required) End port (or ICMP code if protocol is "icmp").
* `type` - (Required) Type of rule being created. Valid options are `ingress` (inbound)
or `egress` (outbound).

The following arguments are optional:

~> **Note** Although `cidr_blocks`, `ipv6_cidr_blocks`, `prefix_list_ids`, and `source_security_group_id` are all marked as optional, you _must_ provide one of them in order to configure the source of the traffic.

* `cidr_blocks` - (Optional) List of CIDR blocks. Cannot be specified with `source_security_group_id` or `self`.
* `description` - (Optional) Description of the rule.
* `ipv6_cidr_blocks` - (Optional) List of IPv6 CIDR blocks. Cannot be specified with `source_security_group_id` or `self`.
* `prefix_list_ids` - (Optional) List of Prefix List IDs.
* `self` - (Optional) Whether the security group itself will be added as a source to this ingress rule. Cannot be specified with `cidr_blocks`, `ipv6_cidr_blocks`, or `source_security_group_id`.
* `source_security_group_id` - (Optional) Security group id to allow access to/from, depending on the `type`. Cannot be specified with `cidr_blocks`, `ipv6_cidr_blocks`, or `self`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the security group rule.
* `security_group_rule_id` - If the `aws_security_group_rule` resource has a single source or destination then this is the AWS Security Group Rule resource ID. Otherwise it is empty.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Group Rules using the `security_group_id`, `type`, `protocol`, `from_port`, `to_port`, and source(s)/destination(s) (such as a `cidr_block`) separated by underscores (`_`). All parts are required. For example:

**NOTE:** Not all rule permissions (e.g., not all of a rule's CIDR blocks) need to be imported for Terraform to manage rule permissions. However, importing some of a rule's permissions but not others, and then making changes to the rule will result in the creation of an additional rule to capture the updated permissions. Rule permissions that were not imported are left intact in the original rule.

Import an ingress rule in security group `sg-6e616f6d69` for TCP port 8000 with an IPv4 destination CIDR of `10.0.3.0/24`:

```terraform
import {
  to = aws_security_group_rule.ingress
  id = "sg-6e616f6d69_ingress_tcp_8000_8000_10.0.3.0/24"
}
```

Import a rule with various IPv4 and IPv6 source CIDR blocks:

```terraform
import {
  to = aws_security_group_rule.ingress
  id = "sg-4973616163_ingress_tcp_100_121_10.1.0.0/16_2001:db8::/48_10.2.0.0/16_2002:db8::/48"
}
```

Import a rule, applicable to all ports, with a protocol other than TCP/UDP/ICMP/ICMPV6/ALL, e.g., Multicast Transport Protocol (MTP), using the IANA protocol number. For example: 92.

```terraform
import {
  to = aws_security_group_rule.ingress
  id = "sg-6777656e646f6c796e_ingress_92_0_65536_10.0.3.0/24_10.0.4.0/24"
}
```

Import a default any/any egress rule to 0.0.0.0/0:

```terraform
import {
  to = aws_security_group_rule.default_egress
  id = "sg-6777656e646f6c796e_egress_all_0_0_0.0.0.0/0"
}
```

Import an egress rule with a prefix list ID destination:

```terraform
import {
  to = aws_security_group_rule.egress
  id = "sg-62726f6479_egress_tcp_8000_8000_pl-6469726b"
}
```

Import a rule applicable to all protocols and ports with a security group source:

```terraform
import {
  to = aws_security_group_rule.ingress_rule
  id = "sg-7472697374616e_ingress_all_0_65536_sg-6176657279"
}
```

Import a rule that has itself and an IPv6 CIDR block as sources:

```terraform
import {
  to = aws_security_group_rule.rule_name
  id = "sg-656c65616e6f72_ingress_tcp_80_80_self_2001:db8::/48"
}
```

**Using `terraform import` to import** Security Group Rules using the `security_group_id`, `type`, `protocol`, `from_port`, `to_port`, and source(s)/destination(s) (such as a `cidr_block`) separated by underscores (`_`). All parts are required. For example:

**NOTE:** Not all rule permissions (e.g., not all of a rule's CIDR blocks) need to be imported for Terraform to manage rule permissions. However, importing some of a rule's permissions but not others, and then making changes to the rule will result in the creation of an additional rule to capture the updated permissions. Rule permissions that were not imported are left intact in the original rule.

Import an ingress rule in security group `sg-6e616f6d69` for TCP port 8000 with an IPv4 destination CIDR of `10.0.3.0/24`:

```console
% terraform import aws_security_group_rule.ingress sg-6e616f6d69_ingress_tcp_8000_8000_10.0.3.0/24
```

Import a rule with various IPv4 and IPv6 source CIDR blocks:

```console
% terraform import aws_security_group_rule.ingress sg-4973616163_ingress_tcp_100_121_10.1.0.0/16_2001:db8::/48_10.2.0.0/16_2002:db8::/48
```

Import a rule, applicable to all ports, with a protocol other than TCP/UDP/ICMP/ICMPV6/ALL, e.g., Multicast Transport Protocol (MTP), using the IANA protocol number. For example: 92.

```console
% terraform import aws_security_group_rule.ingress sg-6777656e646f6c796e_ingress_92_0_65536_10.0.3.0/24_10.0.4.0/24
```

Import a default any/any egress rule to 0.0.0.0/0:

```console
% terraform import aws_security_group_rule.default_egress sg-6777656e646f6c796e_egress_all_0_0_0.0.0.0/0
```

Import an egress rule with a prefix list ID destination:

```console
% terraform import aws_security_group_rule.egress sg-62726f6479_egress_tcp_8000_8000_pl-6469726b
```

Import a rule applicable to all protocols and ports with a security group source:

```console
% terraform import aws_security_group_rule.ingress_rule sg-7472697374616e_ingress_all_0_65536_sg-6176657279
```

Import a rule that has itself and an IPv6 CIDR block as sources:

```console
% terraform import aws_security_group_rule.rule_name sg-656c65616e6f72_ingress_tcp_80_80_self_2001:db8::/48
```
