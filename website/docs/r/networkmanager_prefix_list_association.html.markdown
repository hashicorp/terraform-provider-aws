---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_prefix_list_association"
description: |-
  Associates a prefix list with a Network Manager Cloud WAN core network.
---

# Resource: aws_networkmanager_prefix_list_association

Associates an EC2 managed prefix list with a Network Manager Cloud WAN core network. Once associated, the prefix list can be referenced in the core network policy document.

~> **NOTE:** The prefix list must be defined in the [Cloud WAN home region](https://docs.aws.amazon.com/network-manager/latest/cloudwan/what-is-cloudwan.html#cloudwan-home-region) (us-west-2). Although defined in the Cloud WAN home region, the prefix-list based policy will apply globally to all the relevant core network edges (regions) in your core network.

## Example Usage

```terraform
provider "aws" {
  alias  = "cloudwan_home_region"
  region = "us-west-2"
}

resource "aws_ec2_managed_prefix_list" "prefix_list" {
  provider = aws.cloudwan_home_region

  name           = "example"
  address_family = "IPv4"
  max_entries    = 5

  entry {
    cidr        = "10.0.0.0/8"
    description = "Example CIDR"
  }
}

resource "aws_networkmanager_prefix_list_association" "pl_association" {
  core_network_id   = aws_networkmanager_core_network.core_network.id
  prefix_list_arn   = aws_ec2_managed_prefix_list.prefix_list.arn
  prefix_list_alias = "exampleprefixlist"
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required, Forces new resource) The ID of the core network to associate the prefix list with.
* `prefix_list_alias` - (Required, Forces new resource) An alias for the prefix list association. This alias can be used to reference the prefix list in the core network policy document. Must start with a letter, be less than 64 characters long, and may only include letters and numbers.
* `prefix_list_arn` - (Required, Forces new resource) The ARN of the EC2 managed prefix list to associate with the core network.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networkmanager_prefix_list_association.example
  identity = {
    core_network_id = "core-network-0fab1c1e1e1e1e1e1"
    prefix_list_arn = "arn:aws:ec2:us-west-2:123456789012:prefix-list/pl-0123456789abcdef0"
  }
}

resource "aws_networkmanager_prefix_list_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `core_network_id` (String) Core network ID.
* `prefix_list_arn` (String) Prefix list ARN.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_prefix_list_association` using the core network ID and prefix list ARN separated by a comma (`,`). For example:

```terraform
import {
  to = aws_networkmanager_prefix_list_association.example
  id = "core-network-0fab1c1e1e1e1e1e1,arn:aws:ec2:us-west-2:123456789012:prefix-list/pl-0123456789abcdef0"
}
```

Using `terraform import`, import `aws_networkmanager_prefix_list_association` using the core network ID and prefix list ARN separated by a comma (`,`). For example:

```console
% terraform import aws_networkmanager_prefix_list_association.example core-network-0fab1c1e1e1e1e1e1,arn:aws:ec2:us-west-2:123456789012:prefix-list/pl-0123456789abcdef0
```
