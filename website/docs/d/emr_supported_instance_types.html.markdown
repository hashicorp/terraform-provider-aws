---
subcategory: "EMR"
layout: "aws"
page_title: "AWS: aws_emr_supported_instance_types"
description: |-
  Terraform data source for managing AWS EMR Supported Instance Types.
---

# Data Source: aws_emr_supported_instance_types

Terraform data source for managing AWS EMR Supported Instance Types.

## Example Usage

### Basic Usage

```terraform
data "aws_emr_supported_instance_types" "example" {
  release_label = "ebs-6.15.0"
}
```

### With a Lifecycle Pre-Condition

This data source can be used with a [lifecycle precondition](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#custom-condition-checks) to ensure a given instance type is supported by EMR.

```terraform
locals {
  instance_type = "r7g.large"
  release_label = "emr-6.15.0"
}

data "aws_emr_supported_instance_types" "test" {
  release_label = local.release_label
}

resource "aws_emr_cluster" "test" {
  ### additional configuration omitted for brevity ###

  release_label = local.release_label
  master_instance_group {
    instance_type = local.instance_type
  }

  lifecycle {
    precondition {
      condition     = contains(data.aws_emr_supported_instance_types.test.supported_instance_types[*].type, local.instance_type)
      error_message = "${local.instance_type} is not supported with this EMR release label!"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `release_label` - (Required) Amazon EMR release label. For more information about Amazon EMR releases and their included application versions and features, see the [Amazon EMR Release Guide](https://docs.aws.amazon.com/emr/latest/ReleaseGuide/emr-release-components.html).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `supported_instance_types` - List of supported instance types. See [`supported_instance_types`](#supported_instance_types-attribute-reference) below.

### `supported_instance_types` Attribute Reference

* `architecture` - CPU architecture.
* `ebs_optimized_available` - Indicates whether the instance type supports Amazon EBS optimization.
* `ebs_optimized_by_default` - Indicates whether the instance type uses Amazon EBS optimization by default.
* `ebs_storage_only` - Indicates whether the instance type only supports Amazon EBS.
* `instance_family_id` - The Amazon EC2 family and generation for the instance type.
* `is_64_bits_only` - Indicates whether the instance type only supports 64-bit architecture.
* `memory_gb` - Memory that is available to Amazon EMR from the instance type.
* `number_of_disks` - Number of disks for the instance type.
* `storage_gb` - Storage capacity of the instance type.
* `type` - Amazon EC2 instance type. For example, `m5.xlarge`.
* `vcpu` - The number of vCPUs available for the instance type.
