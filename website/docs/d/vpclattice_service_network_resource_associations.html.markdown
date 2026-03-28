---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service_network_resource_associations"
description: |-
  Terraform data source for managing an AWS VPC Lattice Service Network Resource Associations.
---

# Data Source: aws_vpclattice_service_network_resource_associations

Terraform data source for listing AWS VPC Lattice Service Network Resource Associations.

## Example Usage

### By Service Network Identifier

```terraform
data "aws_vpclattice_service_network_resource_associations" "test_sn" {
  service_network_identifier = aws_vpclattice_service_network.test_sn.id
}
```

### By Resource Configuration Identifier

```terraform
data "aws_vpclattice_service_network_resource_associations" "test_rcfg" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.test.id
}
```

## Argument Reference

This data source supports the following arguments:

* `service_network_identifier` - (Optional) This is the Id or ARN of the VPC Lattice Service Network for which you want to list the Service Network Resource Associations. Use either `service_network_identifier` or `resource_configuration_identifier` but not both.
* `resource_configuration_identifier` - (Optional) This is the Id or ARN of the VPC Lattice Resource Configuration for which you want to list the Service Network Resource Associations. Use either `resource_configuration_identifier` or `service_network_identifier` but not both.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `associations` - List of Objects containing Service Network Resource Associations Summaries ([SDK](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/vpclattice/types#ServiceNetworkResourceAssociationSummary))
    * `arn` -  The Amazon Resource Name (ARN) of the association
    * `created_by` - The account that created the association.
    * `dns_entry` - List of objects with public DNS names.
        * `domain_name` - The domain name of the resource.
        * `hosted_zone_id` - The ID of the hosted zone.
    * `failure_code` - The failure code in cases the association failed.
    * `id` - The ID of the association.
    * `is_managed_association` - Specifies whether the association is managed by Amazon - E.g. when using ARN based resource configurations.
    * `private_dns_entry` - List of objects with private DNS names.
        * `domain_name` - The private domain name of the resource
        * `hosted_zone_id` - The ID of the hosted zone.
    * `resource_configuration_arn` - The ARN of the resource configuration associated with the service network.
    * `resource_configuration_id` - The ID of the resource configuration associated with the service network.
    * `resource_configuration_name` - The name of the resource configuration associated with the service network.
    * `service_network_arn` - The Amazon Resource Name (ARN) of the service network the Service is associated with.
    * `service_network_id` - The ID of the service network the Service is associated with.
    * `service_network_name` - The name of the service network the Service is associated with.
    * `status` - The Status of the Association ("CREATE_IN_PROGRESS", "ACTIVE", "DELETE_IN_PROGRESS", "CREATE_FAILED", "DELETE_FAILED").
