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
data "aws_vpclattice_service_network_service_associations" "test_svc" {
  service_identifier = aws_vpclattice_service.test_svc.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_identifier` - (Optional) ID or ARN of the VPC Lattice Service for which you want to list the Service Network Service Associations. Use either `service_identifier` or `service_network_identifier` but not both.
* `service_network_identifier` - (Optional) ID or ARN of the VPC Lattice Service Network for which you want to list the Service Network Service Associations. Use either `service_network_identifier` or `service_identifier` but not both.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `items` - List of objects containing Service Network Service Associations. Each object has the following attributes:
    * `arn` - ARN of the association.
    * `created_at` - Date and time the association was created, in RFC 3339 format.
    * `created_by` - Account that created the association.
    * `custom_domain_name` - Custom domain name of the service.
    * `dns_entry` - List of objects with DNS names.
        * `domain_name` - Domain name of the service.
        * `hosted_zone_id` - ID of the hosted zone.
    * `id` - ID of the association.
    * `service_arn` - ARN of the associated service.
    * `service_id` - ID of the associated service.
    * `service_name` - Name of the associated service.
    * `service_network_arn` - ARN of the service network the service is associated with.
    * `service_network_id` - ID of the service network the service is associated with.
    * `service_network_name` - Name of the service network the service is associated with.
    * `status` - Status of the association. One of `CREATE_IN_PROGRESS`, `ACTIVE`, `DELETE_IN_PROGRESS`, `CREATE_FAILED`, or `DELETE_FAILED`.
