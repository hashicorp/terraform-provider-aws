---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service_network"
description: |-
  Terraform data source for managing an AWS VPC Lattice Service Network.
---

# Data Source: aws_vpclattice_service_network

Terraform data source for managing an AWS VPC Lattice Service Network.

## Example Usage

### Basic Usage

```terraform
data "aws_vpclattice_service_network" "example" {
  service_network_identifier = ""
}
```

## Argument Reference

The following arguments are required:

* `service_network_identifier` - (Required) Identifier of the network service.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Network.
* `auth_type` - Authentication type for the service network. Either `NONE` or `AWS_IAM`.
* `created_at` - Date and time the service network was created.
* `id` - ID of the service network.
* `last_updated_at` - Date and time the service network was last updated.
* `name` - Name of the service network.
* `number_of_associated_services` - Number of services associated with this service network.
* `number_of_associated_vpcs` - Number of VPCs associated with this service network.
