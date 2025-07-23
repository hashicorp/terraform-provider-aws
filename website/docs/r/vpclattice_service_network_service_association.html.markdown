---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service_network_service_association"
description: |-
  Terraform resource for managing an AWS VPC Lattice Service Network Service Association.
---

# Resource: aws_vpclattice_service_network_service_association

Terraform resource for managing an AWS VPC Lattice Service Network Service Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_service_network_service_association" "example" {
  service_identifier         = aws_vpclattice_service.example.id
  service_network_identifier = aws_vpclattice_service_network.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_identifier` - (Required) The ID or Amazon Resource Identifier (ARN) of the service.
* `service_network_identifier` - (Required) The ID or Amazon Resource Identifier (ARN) of the service network. You must use the ARN if the resources specified in the operation are in different accounts.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Association.
* `created_by` - The account that created the association.
* `custom_domain_name` - The custom domain name of the service.
* `dns_entry` - The DNS name of the service.
    * `domain_name` - The domain name of the service.
    * `hosted_zone_id` - The ID of the hosted zone.
* `id` - The ID of the association.
* `status` - The operations status. Valid Values are CREATE_IN_PROGRESS | ACTIVE | DELETE_IN_PROGRESS | CREATE_FAILED | DELETE_FAILED
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Service Network Service Association using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_service_network_service_association.example
  id = "snsa-05e2474658a88f6ba"
}
```

Using `terraform import`, import VPC Lattice Service Network Service Association using the `id`. For example:

```console
% terraform import aws_vpclattice_service_network_service_association.example snsa-05e2474658a88f6ba
```
