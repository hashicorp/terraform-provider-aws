---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_managed_prefix_list"
description: |-
  Provides a managed prefix list resource.
---

# Resource: aws_ec2_managed_prefix_list

Provides a managed prefix list resource.

~> **NOTE on Managed Prefix Lists and Managed Prefix List Entries:** Terraform
currently provides both a standalone [Managed Prefix List Entry resource](ec2_managed_prefix_list_entry.html) (a single entry),
and a Managed Prefix List resource with entries defined in-line. At this time you
cannot use a Managed Prefix List with in-line rules in conjunction with any Managed
Prefix List Entry resources. Doing so will cause a conflict of entries and will overwrite entries.

~> **NOTE on `max_entries`:** When you reference a Prefix List in a resource,
the maximum number of entries for the prefix lists counts as the same number of rules
or entries for the resource. For example, if you create a prefix list with a maximum
of 20 entries and you reference that prefix list in a security group rule, this counts
as 20 rules for the security group.

## Example Usage

Basic usage

```terraform
resource "aws_ec2_managed_prefix_list" "example" {
  name           = "All VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5

  entry {
    cidr        = aws_vpc.example.cidr_block
    description = "Primary"
  }

  entry {
    cidr        = aws_vpc_ipv4_cidr_block_association.example.cidr_block
    description = "Secondary"
  }

  tags = {
    Env = "live"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `address_family` - (Required, Forces new resource) Address family (`IPv4` or `IPv6`) of this prefix list.
* `entry` - (Optional) Configuration block for prefix list entry. Detailed below. Different entries may have overlapping CIDR blocks, but a particular CIDR should not be duplicated.
* `max_entries` - (Required) Maximum number of entries that this prefix list can contain.
* `name` - (Required) Name of this resource. The name must not start with `com.amazonaws`.
* `tags` - (Optional) Map of tags to assign to this resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `entry`

* `cidr` - (Required) CIDR block of this entry.
* `description` - (Optional) Description of this entry. Due to API limitations, updating only the description of an existing entry requires temporarily removing and re-adding the entry.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the prefix list.
* `id` - ID of the prefix list.
* `owner_id` - ID of the AWS account that owns this prefix list.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - Latest version of this prefix list.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Prefix Lists using the `id`. For example:

```terraform
import {
  to = aws_ec2_managed_prefix_list.default
  id = "pl-0570a1d2d725c16be"
}
```

Using `terraform import`, import Prefix Lists using the `id`. For example:

```console
% terraform import aws_ec2_managed_prefix_list.default pl-0570a1d2d725c16be
```
