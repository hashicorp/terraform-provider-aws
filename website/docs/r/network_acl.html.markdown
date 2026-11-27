---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_acl"
description: |-
  Provides an network ACL resource.
---

# Resource: aws_network_acl

Provides an network ACL resource. You might set up network ACLs with rules similar
to your security groups in order to add an additional layer of security to your VPC.

~> **NOTE on Network ACLs and Network ACL Rules:** Terraform currently
provides both a standalone [Network ACL Rule](network_acl_rule.html) resource and a Network ACL resource with rules
defined in-line. At this time you cannot use a Network ACL with in-line rules
in conjunction with any Network ACL Rule resources. Doing so will cause
a conflict of rule settings and will overwrite rules.

~> **NOTE on Network ACLs and Network ACL Associations:** Terraform provides both a standalone [network ACL association](network_acl_association.html)
resource and a network ACL resource with a `subnet_ids` attribute. Do not use the same subnet ID in both a network ACL
resource and a network ACL association resource. Doing so will cause a conflict of associations and will overwrite the association.

## Example Usage

```terraform
resource "aws_network_acl" "main" {
  vpc_id = aws_vpc.main.id

  egress {
    protocol   = "tcp"
    rule_no    = 200
    action     = "allow"
    cidr_block = "10.3.0.0/18"
    from_port  = 443
    to_port    = 443
  }

  ingress {
    protocol   = "tcp"
    rule_no    = 100
    action     = "allow"
    cidr_block = "10.3.0.0/18"
    from_port  = 80
    to_port    = 80
  }

  tags = {
    Name = "main"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `egress` - (Optional) Egress rule. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
* `ingress` - (Optional) Ingress rule. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subnet_ids` - (Optional) List of Subnet IDs to apply the ACL to.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_id` - (Required) ID of the associated VPC.

### `egress` and `ingress` Block

Both arguments are processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

Both `egress` and `ingress` support the following keys:

* `action` - (Required) Action to take.
* `cidr_block` - (Optional) CIDR block to match. This must be a valid network mask.
* `from_port` - (Required) From port to match.
* `icmp_code` - (Optional) ICMP type code to be used. Default 0.
* `icmp_type` - (Optional) ICMP type to be used. Default 0.
* `ipv6_cidr_block` - (Optional) IPv6 CIDR block.
* `protocol` - (Required) Protocol to match. If using the -1 'all' protocol, you must specify a from and to port of 0.
* `rule_no` - (Required) Rule number. Used for ordering.
* `to_port` - (Required) To port to match.

~> Note: For more information on ICMP types and codes, see here: https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the network ACL
* `id` - ID of the network ACL
* `owner_id` - ID of the AWS account that owns the network ACL.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_network_acl.main
  identity = {
    id = "acl-7aaabd18"
  }
}
```

### Identity Schema

#### Required

* `id` (String) ID of the Network ACL.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network ACLs using the `id`. For example:

```terraform
import {
  to = aws_network_acl.main
  id = "acl-7aaabd18"
}
```

Using `terraform import`, import Network ACLs using the `id`. For example:

```console
% terraform import aws_network_acl.main acl-7aaabd18
```
