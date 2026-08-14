---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_network"
page_title: "AWS: aws_odb_network"
description: |-
  Terraform data source to retrieve odb network for Oracle Database@AWS.
---

# Data Source: aws_odb_network

Terraform data source for to retrieve network resource in AWS for Oracle Database@AWS.

## Example Usage

### Basic Usage

```terraform

data "aws_odb_network" "example" {
  id = "example"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required)  Unique identifier of the odb network resource.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the odb network resource.
* `availability_zone` - Availability zone where the ODB network is located.
* `availability_zone_id` - AZ ID of the AZ where the ODB network is located.
* `backup_subnet_cidr` - CIDR range of the backup subnet for the ODB network.
* `client_subnet_cidr` - CIDR notation for the network resource.
* `created_at` - Date and time when the ODB network was created.
* `custom_domain_name` - Name of the custom domain that the network is located.
* `default_dns_prefix` - Default DNS prefix for the network resource.
* `display_name` - Display name for the network resource.
* `ec2_placement_group_ids` - List of EC2 placement group IDs associated with the ODB network.
* `id` - Unique identifier of the odb network resource.
* `managed_services` - Managed services configuration for the ODB network.
* `oci_dns_forwarding_configs` - DNS resolver endpoint in OCI for forwarding DNS queries for the ociPrivateZone domain.
* `oci_network_anchor_id` - Unique identifier of the OCI network anchor for the ODB network.
* `oci_network_anchor_url` - URL of the OCI network anchor for the ODB network.
* `oci_resource_anchor_name` - Name of the OCI resource anchor for the ODB network.
* `oci_vcn_id` - Unique identifier Oracle Cloud ID (OCID) of the OCI VCN for the ODB network.
* `oci_vcn_url` - URL of the OCI VCN for the ODB network.
* `peered_cidrs` - List of CIDR ranges from the peered VPC that are allowed access to the ODB network. Please refer odb network peering documentation.
* `percent_progress` - Amount of progress made on the current operation on the ODB network, expressed as a percentage.
* `status` - Status of the network resource.
* `status_reason` - Additional information about the current status of the ODB network.
* `tags` - Map of tags assigned to the resource.
