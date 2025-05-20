---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service_network_service_associations"
description: |-
  Terraform data source for listing AWS VPC Lattice Service Network Service Associations.
---

# Data Source: aws_vpclattice_service_network_service_associations

Terraform data source for listing AWS VPC Lattice Service Network Service Associations.

## Example Usage

### By Service Network Identifier

```terraform
data "aws_vpclattice_service_network_service_associations" "test_sn" {
  service_network_identifier = aws_vpclattice_service_network.test_sn.id
}
```

### By Service Identifier

```terraform
data "aws_vpclattice_service_network_service_associations" "test_svc {
  service_identifier = aws_vpclattice_service.test_svc.id
}
```

## Argument Reference

One, and only one of the following arguments is required:

* `service_network_identifier` - (Optional) This is the Id or ARN of the VPC Lattice Service Network for which you want to list the Service Network Service Associations. Use either `service_network_identifier` or `service_identifier` but not both.

* `service_identifier` - (Optional) This is the Id or ARN of the VPC Lattice Service for which you want to list the Service Network Service Associations. Use either `service_identifier` or `service_network_identifier` but not both.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `associations` - List of Objects containing Service Network Service Associations Summaries ([SDK](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/vpclattice@v1.14.2/types#ServiceNetworkServiceAssociationSummary))
  * `arn` -  The Amazon Resource Name (ARN) of the association
  * `created_by` - The account that created the association.
  * `custom_domain_name` - The custom domain name of the service.
  * `dns_entry` - List of objects with DNS names.
    * `domain_name` - The domain name of the service
    * `hosted_zone_id` - The ID of the hosted zone.
  * `id` - The ID of the association.
  * `service_arn` - The Amazon Resource Name (ARN) of the associated service.
  * `service_id` - The ID of the associated service.
  * `service_name` - The name of the associated service.
  * `service_network_arn` - The Amazon Resource Name (ARN) of the service network the Service is associated with.
  * `service_network_id` - The ID of the service network the Service is associated with.
  * `service_network_name` - The name of the service network the Service is associated with.
  * `status` - The Status of the Association ("CREATE_IN_PROGRESS", "ACTIVE", "DELETE_IN_PROGRESS", "CREATE_FAILED", "DELETE_FAILED")

