---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_prefix_list_entry"
description: |-
  Provides a managed prefix list entry resource.
---

# Resource: aws_prefix_list_entry

Provides a managed prefix list entry resource. Represents a single `entry`, which
can be added to external Prefix Lists.

~> **NOTE on Prefix Lists and Prefix List Entries:** Terraform currently
provides both a standalone Prefix List Entry, and a [Prefix List resource](prefix_list.html) 
with an `entry` set defined in-line. At this time you
cannot use a Prefix List with in-line rules in conjunction with any Prefix List Entry
resources. Doing so will cause a conflict of rule settings and will unpredictably
fail or overwrite rules.

~> **NOTE:** A Prefix List will have an upper bound on the number of rules
that it can support.

~> **NOTE:** Resource creation will fail if the target Prefix List already has a
rule against the given CIDR block.

## Example Usage

Basic usage

```hcl
resource "aws_prefix_list" "example" {
  name           = "All VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_prefix_list_entry" "example" {
  prefix_list_id = aws_prefix_list.example.id
  cidr_block     = aws_vpc.example.cidr_block
  description    = "Primary"
}
```

## Argument Reference

The following arguments are supported:

* `prefix_list_id` - (Required, Forces new resource) ID of the Prefix List to add this entry to.
* `cidr_block` - (Required, Forces new resource) The CIDR block to add an entry for. Different entries may have
    overlapping CIDR blocks, but duplicating a particular block is not allowed.
* `description` - (Optional, Up to 255 characters) The description of this entry. 

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the prefix list entry.

## Import

Prefix List Entries can be imported using a concatenation of the `prefix_list_id` and `cidr_block` by an underscore (`_`). For example:

```console
$ terraform import aws_prefix_list_entry.example pl-0570a1d2d725c16be_10.30.0.0/16
```
