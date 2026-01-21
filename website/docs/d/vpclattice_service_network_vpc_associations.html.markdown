---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service_network_vpc_associations"
description: |-
  Terraform data source for managing an AWS VPC Lattice Service Network VPC Associations.
---

# Data Source: aws_vpclattice_service_network_vpc_associations

Terraform data source for listing AWS VPC Lattice Service Network VPC Associations.

## Example Usage

### By Service Network Identifier

```terraform
data "aws_vpclattice_service_network_vpc_associations" "test_sn" {
  service_network_identifier = aws_vpclattice_service_network.test_sn.id
}
```

### By VPC Identifier

```terraform
data "aws_vpclattice_service_network_vpc_associations" "test_vpc" {
  vpc_identifier = aws_vpc.test_vpc.id
}
```

## Argument Reference

This data source supports the following arguments:

* `service_network_identifier` - (Optional) This is the Id or ARN of the VPC Lattice Service Network for which you want to list the Service Network VPC Associations. Use either `service_network_identifier` or `vpc_identifier` but not both.
* `vpc_identifier` - (Optional) This is the Id of the VPC for which you want to list the Service Network VPC Associations. Use either `vpc_identifier` or `service_network_identifier` but not both.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `associations` - List of Objects containing Service Network VPC Associations Summaries ([SDK](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/vpclattice/types#ServiceNetworkVPCAssociationSummary))
    * `arn` -  The Amazon Resource Name (ARN) of the association
    * `created_by` - The account that created the association.
    * `id` - The ID of the association.
    * `service_network_arn` - The Amazon Resource Name (ARN) of the service network the VPC is associated with.
    * `service_network_id` - The ID of the service network the VPC is associated with.
    * `service_network_name` - The name of the service network the VPC is associated with.
    * `status` - The Status of the Association ("CREATE_IN_PROGRESS", "ACTIVE", "DELETE_IN_PROGRESS", "CREATE_FAILED", "DELETE_FAILED", "UPDATE_FAILED")
    * `vpc_id`- The Id of the associated VPC
