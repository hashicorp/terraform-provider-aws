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

* `arn` - ARN of the odb network resource.
* `availability_zone` - Availability zone where the ODB network is located.
* `availability_zone_id` - AZ ID of the AZ where the ODB network is located.
* `backup_subnet_cidr` - CIDR range of the backup subnet for the ODB network.
* `client_subnet_cidr` - CIDR notation for the network resource.
* `created_at` - Date and time when the ODB network was created.
* `custom_domain_name` - Name of the custom domain that the network is located.
* `default_dns_prefix` - Default DNS prefix for the network resource.
* `display_name` - Display name for the network resource.
* `ec2_placement_group_ids` - List of EC2 placement group IDs associated with the ODB network.
* `managed_services` - Managed services configuration for the ODB network. See [`managed_services` Block](#managed_services-block) below.
* `oci_dns_forwarding_configs` - DNS resolver endpoint in OCI for forwarding DNS queries for the ociPrivateZone domain. See [`oci_dns_forwarding_configs` Block](#oci_dns_forwarding_configs-block) below.
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

### `managed_services` Block

* `cross_region_s3_restore_sources_access` - Cross-region S3 restore sources access configuration. See [`cross_region_s3_restore_sources_access` Block](#cross_region_s3_restore_sources_access-block) below.
* `kms_access` - Amazon KMS access configuration. See [`kms_access` Block](#kms_access-block) below.
* `managed_s3_backup_access` - Managed S3 backup access configuration. See [`managed_s3_backup_access` Block](#managed_s3_backup_access-block) below.
* `managed_service_ipv4_cidrs` - List of IPv4 CIDR ranges used by the managed services.
* `resource_gateway_arn` - ARN of the resource gateway.
* `s3_access` - Amazon S3 access configuration. See [`s3_access` Block](#s3_access-block) below.
* `service_network_arn` - ARN of the service network.
* `service_network_endpoint` - Service network endpoint configuration. See [`service_network_endpoint` Block](#service_network_endpoint-block) below.
* `sts_access` - Amazon STS access configuration. See [`sts_access` Block](#sts_access-block) below.
* `zero_tl_access` - Zero-ETL access configuration. See [`zero_tl_access` Block](#zero_tl_access-block) below.

#### `cross_region_s3_restore_sources_access` Block

* `ipv4_addresses` - List of IPv4 addresses for the cross-region S3 restore sources access.
* `region` - Region enabled for cross-region S3 restore.
* `status` - Status of the cross-region S3 restore sources access.

#### `kms_access` Block

* `domain_name` - Domain name for the Amazon KMS access.
* `ipv4_addresses` - List of IPv4 addresses for the Amazon KMS access.
* `kms_policy_document` - Endpoint policy for the Amazon KMS access.
* `status` - Status of the Amazon KMS access.

#### `managed_s3_backup_access` Block

* `ipv4_addresses` - List of IPv4 addresses for the managed S3 backup access.
* `status` - Status of the managed S3 backup access.

#### `s3_access` Block

* `domain_name` - Domain name for the Amazon S3 access.
* `ipv4_addresses` - List of IPv4 addresses for the Amazon S3 access.
* `s3_policy_document` - Endpoint policy for the Amazon S3 access.
* `status` - Status of the Amazon S3 access.

#### `service_network_endpoint` Block

* `vpc_endpoint_id` - Unique identifier of the VPC endpoint.
* `vpc_endpoint_type` - Type of the VPC endpoint.

#### `sts_access` Block

* `domain_name` - Domain name for the Amazon STS access.
* `ipv4_addresses` - List of IPv4 addresses for the Amazon STS access.
* `status` - Status of the Amazon STS access.
* `sts_policy_document` - Endpoint policy for the Amazon STS access.

#### `zero_tl_access` Block

* `cidr` - CIDR range for the Zero-ETL access.
* `status` - Status of the Zero-ETL access.

### `oci_dns_forwarding_configs` Block

* `domain_name` - Domain name for which the DNS queries are forwarded.
* `oci_dns_listener_ip` - IP address of the OCI DNS listener.
