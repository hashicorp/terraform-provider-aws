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
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) A description for the IPAM Resource Discovery Association.
* `ipam_id` - (Required) Id of the IPAM to associate
* `resource_discovery_id` - (Required) Id of the Resource Discovery to associate

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of IPAM Resource Discovery Association
* `id` - The ID of the IPAM Resource Discovery Association
* `is_default` - A boolean to identify if the Resource Discovery is the accounts default resource discovery
* `owner_id` - The account ID for the account that manages the Resource Discovery
* `ipam_resource_discovery_region` - The home region of the Resource Discovery Association
* `ipam_region` - The home region of the IPAM
* `ipam_arn` - The arn of the IPAM
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

IPAMs can be imported using the `ipam resource discovery association id`, e.g.

```
$ terraform import aws_vpc_ipam_resource_discovery_association.example ipam-res-disco-assoc-0178368ad2146a492
```
