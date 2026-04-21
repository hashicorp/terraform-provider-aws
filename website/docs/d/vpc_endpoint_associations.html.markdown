---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_associations"
description: |-
  Provides details of Resource and Service Network associations to a VPC Endpoint.
---

# Data Source: aws_vpc_endpoint_associations

Terraform data source for managing an AWS EC2 (Elastic Compute Cloud) Vpc Endpoint Associations.

## Example Usage

### Basic Usage

```terraform
data "aws_vpc_endpoint_associations" "example" {
  vpc_endpoint_id = aws_vpc_endpoint.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_endpoint_id` - ID of the specific VPC Endpoint to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `associations` - Associations for the VPC Endpoint. [Association blocks are documented below](#associations-block).

### `associations` Block

Associations blocks (for `associations`) support the following attributes:

* `associated_resource_accessibility` - Accessibility of the resource.
* `associated_resource_arn` - ARN of the resource for this association.
* `dns_entry` - DNS entries for the Association. [DNS entry blocks are documented below](#dns_entry-block).
* `private_dns_entry` - DNS entries for the Association. [Private DNS entry blocks are documented below](#private_dns_entry-block).
* `resource_configuration_group_arn` - ARN of the Resource Group if the Resource is a member of a group.
* `service_network_arn` - Service Network ARN. Applicable for endpoints of type `ServiceNetwork`.
* `service_network_name` - Service Network Name. Applicable for endpoints of type `ServiceNetwork`.
* `tags` - Tags of the association.

### `dns_entry` Block

DNS blocks (for `dns_entry`) support the following attributes:

* `dns_name` - DNS name.
* `hosted_zone_id` - ID of the private hosted zone.

### `private_dns_entry` Block

DNS blocks (for `private_dns_entry`) support the following attributes:

* `dns_name` - DNS name.
* `hosted_zone_id` - ID of the private hosted zone.
