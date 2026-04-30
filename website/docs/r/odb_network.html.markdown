---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_network"
page_title: "AWS: aws_odb_network"
description: |-
  Terraform resource for managing odb network of an Oracle Database@AWS.
---

# Resource: aws_odb_network

Terraform resource for managing odb Network resource in AWS for Oracle Database@AWS.

## Example Usage

### Basic Usage

```terraform

resource "aws_odb_network" "example" {
  display_name         = "odb-my-net"
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
  tags = {
    "env" = "dev"
  }
}

resource "aws_odb_network" "example" {
  display_name         = "odb-my-net"
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "ENABLED"
  zero_etl_access      = "ENABLED"
  tags = {
    "env" = "dev"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) The user-friendly name for the odb network. Changing this will force terraform to create a new resource.
* `availability_zone_id` - (Required) The AZ ID of the AZ where the ODB network is located. Changing this will force terraform to create new resource.
* `client_subnet_cidr` - (Required) The CIDR notation for the network resource. Changing this will force terraform to create new resource.
* `backup_subnet_cidr` - (Required) The CIDR range of the backup subnet for the ODB network. Changing this will force terraform to create new resource.
* `s3_access` - (Required) Specifies the configuration for Amazon S3 access from the ODB network.
* `zero_etl_access` - (Required) Specifies the configuration for Zero-ETL access from the ODB network.

The following arguments are optional:

* `custom_domain_name` - (Optional) The name of the custom domain that the network is located. Custom_domain_name and default_dns_prefix both can't be given. Changing this will force terraform to create new resource.
* `availability_zone` - (Optional) The name of the Availability Zone (AZ) where the odb network is located. Changing this will force terraform to create new resource. Make sure availability_zone maps correctly with availability_zone_id.
* `s3_policy_document` - (Optional) Specifies the endpoint policy for Amazon S3 access from the ODB network.
* `default_dns_prefix` - (Optional) The default DNS prefix for the network resource. Changing this will force terraform to create new resource. Changing this will force terraform to create new resource.
* `tags` - (Optional) A map of tags to assign to the exadata infrastructure. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `delete_associated_resources` - (Optional) If set to true deletes associated OCI resources. Default false.
* `sts_access` - (Optional) Specifies the configuration for STS access from the ODB network.
* `kms_access` - (Optional) Specifies the configuration for KMS access from the ODB network.
* `sts_policy_document` - (Optional) Specifies the endpoint policy for STS access from the ODB network.
* `kms_policy_document` - (Optional) Specifies the endpoint policy for KMS access from the ODB network.
* `cross_region_s3_restore_sources_access` - (Optional) The list of regions enabled for cross-region restore in the ODB network.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier of the odb network resource.
* `arn` - Amazon Resource Name (ARN) of the odb network resource.
* `oci_dns_forwarding_configs` - The number of storage servers requested for the Exadata infrastructure.
* `peered_cidrs` - The list of CIDR ranges from the peered VPC that are allowed access to the ODB network. Please refer odb network peering documentation.
* `oci_network_anchor_id` - The unique identifier of the OCI network anchor for the ODB network.
* `oci_network_anchor_url` -The URL of the OCI network anchor for the ODB network.
* `oci_resource_anchor_name` - The name of the OCI resource anchor for the ODB network.
* `oci_vcn_id` - The unique identifier  Oracle Cloud ID (OCID) of the OCI VCN for the ODB network.
* `oci_vcn_url` - The URL of the OCI VCN for the ODB network.
* `percent_progress` - The amount of progress made on the current operation on the ODB network, expressed as a percentage.
* `managed_services` - The managed services configuration for the ODB network.
* `status` - The status of the network resource.
* `status_reason` - Additional information about the current status of the ODB network.
* `created_at` - The date and time when the ODB network was created.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `24h`)
* `update` - (Default `24h`)
* `delete` - (Default `24h`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Ingestion Pipeline using the `id`. For example:

```terraform
import {
  to = aws_odb_network.example
  id = "example"
}
```

Using `terraform import`, import Odb Network using the `id`. For example:

```console
% terraform import aws_odb_network.example example
```
