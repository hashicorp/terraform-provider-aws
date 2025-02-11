---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service_network_resource_association"
description: |-
  Terraform resource for managing an AWS VPC Lattice Service Network Resource Association.
---
# Resource: aws_vpclattice_service_network_resource_association

Terraform resource for managing an AWS VPC Lattice Service Network Resource Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_service_network_resource_association" "example" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.example.id
  service_network_identifier        = aws_vpclattice_service_network.example.id

  tags = {
    Name = "Example"
  }
}
```

## Argument Reference

The following arguments are required:

* `resource_configuration_identifier` - (Required) Identifier of Resource Configuration to associate to the Service Network.
* `service_network_identifier` - (Required) Identifier of the Service Network to associate the Resource to.

The following arguments are optional:

* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Network Resource Association.
* `id` - ID of the association.
* `dns_entry` DNS entry of the association in the service network.
    * `domain_name` The domain name of the association in the service network.
    * `hosted_zone_id` The ID of the hosted zone containing the domain name.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Service Network Resource Association using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_service_network_resource_association.example
  id = "snra-1234567890abcef12"
}
```

Using `terraform import`, import VPC Lattice Service Network Resource Association using the `id`. For example:

```console
% terraform import aws_vpclattice_service_network_resource_association.example snra-1234567890abcef12
```
