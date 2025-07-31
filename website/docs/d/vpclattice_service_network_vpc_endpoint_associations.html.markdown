---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service_network_vpc_endpoint_associations"
description: |-
  Terraform data source for listing AWS VPC Lattice Service Network VPC Endpoint Associations.
---

# Data Source: aws_vpclattice_service_network_vpc_endpoint_associations

Terraform data source for listing AWS VPC Lattice Service Network VPC Endpoint Associations.

## Example Usage

### Basic Usage

```terraform
data "aws_vpclattice_service_network_vpc_endpoint_associations" "example" {
  service_network_identifier = aws_vpclattice_service_network.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `service_network_identifier` - (Required) This is the Id or ARN of the VPC Lattice Service Network for which you want to list the Service Network VPC Endpoint Associations.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `associations` - List of Objects containing Service Network VPC Endpoint Associations ([SDK](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/vpclattice/types#ServiceNetworkEndpointAssociation))
    * `id` - The ID of the association.
    * `service_network_arn` - The Amazon Resource Name (ARN) of the service network the VPC Endpoint is associated with.
    * `state` - The State of the Association
    * `vpc_endpoint_id` - The ID of the VPC Endpoint associated with the service network.
    * `vpc_endpoint_owner_id` - The owner account of the VPC Endpoint associated with the service network.
    * `vpc_id` - The ID of the VPC for the association.
