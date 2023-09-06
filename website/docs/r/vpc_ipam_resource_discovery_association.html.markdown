---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_resource_discovery_association"
description: |-
  Provides an IPAM Resource Discovery Association resource.
---

# Resource: aws_vpc_ipam_resource_discovery_association

Provides an association between an Amazon IP Address Manager (IPAM) and a IPAM Resource Discovery. IPAM Resource Discoveries are resources meant for multi-organization customers. If you wish to use a single IPAM across multiple orgs, a resource discovery can be created and shared from a subordinate organization to the management organizations IPAM delegated admin account.

Once an association is created between two organizations via IPAM & a IPAM Resource Discovery, IPAM Pools can be shared via Resource Access Manager (RAM) to accounts in the subordinate organization; these RAM shares must be accepted by the end user account. Pools can then also discover and monitor IPAM resources in the subordinate organization.

## Example Usage

Basic usage:

```terraform
resource "aws_vpc_ipam_resource_discovery_association" "test" {
  ipam_id                    = aws_vpc_ipam.test.id
  ipam_resource_discovery_id = aws_vpc_ipam_resource_discovery.test.id

  tags = {
    "Name" = "test"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `ipam_id` - (Required) The ID of the IPAM to associate.
* `ipam_resource_discovery_id` - (Required) The ID of the Resource Discovery to associate.
* `tags` - (Optional) A map of tags to add to the IPAM resource discovery association resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of IPAM Resource Discovery Association.
* `id` - The ID of the IPAM Resource Discovery Association.
* `owner_id` - The account ID for the account that manages the Resource Discovery
* `ipam_arn` - The Amazon Resource Name (ARN) of the IPAM.
* `ipam_region` - The home region of the IPAM.
* `is_default` - A boolean to identify if the Resource Discovery is the accounts default resource discovery.
* `state` - The lifecycle state of the association when you associate or disassociate a resource discovery.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IPAMs using the IPAM resource discovery association `id`. For example:

```terraform
import {
  to = aws_vpc_ipam_resource_discovery_association.example
  id = "ipam-res-disco-assoc-0178368ad2146a492"
}
```

Using `terraform import`, import IPAMs using the IPAM resource discovery association `id`. For example:

```console
% terraform import aws_vpc_ipam_resource_discovery_association.example ipam-res-disco-assoc-0178368ad2146a492
```
