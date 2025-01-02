---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_managed_prefix_list_entry"
description: |-
  Use the `aws_ec2_managed_prefix_list_entry` resource to manage a managed prefix list entry.
---

# Resource: aws_ec2_managed_prefix_list_entry

Use the `aws_prefix_list_entry` resource to manage a managed prefix list entry.

~> **NOTE:** Terraform currently provides two resources for managing Managed Prefix Lists and Managed Prefix List Entries. The standalone resource, [Managed Prefix List Entry](ec2_managed_prefix_list_entry.html), is used to manage a single entry. The [Managed Prefix List resource](ec2_managed_prefix_list.html) is used to manage multiple entries defined in-line. It is important to note that you cannot use a Managed Prefix List with in-line rules in conjunction with any Managed Prefix List Entry resources. This will result in a conflict of entries and will cause the entries to be overwritten.

~> **NOTE:** To improve execution times on larger updates, it is recommended to use the inline `entry` block as part of the Managed Prefix List resource when creating a prefix list with more than 100 entries. You can find more information about the resource [here](ec2_managed_prefix_list.html).

## Example Usage

Basic usage.

```terraform
resource "aws_ec2_managed_prefix_list" "example" {
  name           = "All VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    Env = "live"
  }
}

resource "aws_ec2_managed_prefix_list_entry" "entry_1" {
  cidr           = aws_vpc.example.cidr_block
  description    = "Primary"
  prefix_list_id = aws_ec2_managed_prefix_list.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `cidr` - (Required) CIDR block of this entry.
* `description` - (Optional) Description of this entry. Please note that due to API limitations, updating only the description of an entry will require recreating the entry.
* `prefix_list_id` - (Required) The ID of the prefix list.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the managed prefix list entry.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import prefix list entries using `prefix_list_id` and `cidr` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ec2_managed_prefix_list_entry.default
  id = "pl-0570a1d2d725c16be,10.0.3.0/24"
}
```

Using `terraform import`, import prefix list entries using `prefix_list_id` and `cidr` separated by a comma (`,`). For example:

```console
% terraform import aws_ec2_managed_prefix_list_entry.default pl-0570a1d2d725c16be,10.0.3.0/24
```
