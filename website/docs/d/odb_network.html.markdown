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

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier of the odb network resource.
* `arn` - Amazon Resource Name (ARN) of the odb network resource.
* `display_name` - Display name for the network resource.
* `availability_zone_id` - The AZ ID of the AZ where the ODB network is located.
* `availability_zone` - The availability zone where the ODB network is located.
* `backup_subnet_cidr` - The CIDR range of the backup subnet for the ODB network.
* `client_subnet_cidr` - The CIDR notation for the network resource.
* `custom_domain_name` - The name of the custom domain that the network is located.
* `default_dns_prefix` - The default DNS prefix for the network resource.
* `oci_network_anchor_id` - The unique identifier of the OCI network anchor for the ODB network.
* `oci_network_anchor_url` - The URL of the OCI network anchor for the ODB network.
* `oci_resource_anchor_name` - The name of the OCI resource anchor for the ODB network.
* `oci_vcn_id` - The unique identifier  Oracle Cloud ID (OCID) of the OCI VCN for the ODB network.
* `oci_vcn_url` - The URL of the OCI VCN for the ODB network.
* `percent_progress` - The amount of progress made on the current operation on the ODB network, expressed as a percentage.
* `peered_cidrs` - The list of CIDR ranges from the peered VPC that are allowed access to the ODB network. Please refer odb network peering documentation.
* `status` - The status of the network resource.
* `status_reason` - Additional information about the current status of the ODB network.
* `created_at` - The date and time when the ODB network was created.
* `managed_services` - The managed services configuration for the ODB network.
