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

### By Identifier

```terraform
data "aws_vpclattice_service_network" "example" {
  service_network_identifier = "snsa-01112223334445556"
}
```

### By Name

```terraform
data "aws_vpclattice_service_network" "example" {
  name = "my-service-network"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Optional) Name of the service network.
* `service_network_identifier` - (Optional) ID or ARN of the service network.

~> **NOTE:** One of `name` or `service_network_identifier` must be specified.

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
