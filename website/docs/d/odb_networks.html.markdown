---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_networks"
page_title: "AWS: aws_odb_networks"
description: |-
  Terraform data source to odb networks for Oracle Database@AWS.
---

# Data Source: aws_odb_networks

Terraform data source for to retrieve networks from AWS for Oracle Database@AWS.

## Example Usage

### Basic Usage

```terraform

data "aws_odb_networks" "example" {}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `odb_networks` - List of odb networks returns basic information about odb networks.

### odb_networks

* `id` - Unique identifier of the odb network resource.
* `arn` - Amazon Resource Name (ARN) of the odb network resource.
* `oci_network_anchor_id` - The unique identifier of the OCI network anchor for the ODB network.
* `oci_vcn_url` - The URL of the OCI VCN for the ODB network.
* `oci_vcn_id` - The unique identifier  Oracle Cloud ID (OCID) of the OCI VCN for the ODB network.
* `display_name` - Display name for the network resource.
