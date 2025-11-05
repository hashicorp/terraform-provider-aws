---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_db_system_shapes"
page_title: "AWS: aws_odb_db_system_shapes"
description: |-
  Terraform data source to retrieve available system shapes Oracle Database@AWS.
---

# Data Source: aws_odb_db_system_shapes

Terraform data source to retrieve available system shapes Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_db_system_shapes" "example" {}
```

## Argument Reference

The following arguments are optional:

* `availability_zone_id` - (Optional) The physical ID of the AZ, for example, use1-az4. This ID persists across accounts.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `db_system_shapes` - The list of shapes and their properties. Information about a hardware system model (shape) that's available for an Exadata infrastructure. The shape determines resources, such as CPU cores, memory, and storage, to allocate to the Exadata infrastructure.

### db_system_shapes

* `are_server_types_supported` - Indicates whether the hardware system model supports configurable database and server storage types.
* `available_core_count` - The maximum number of CPU cores that can be enabled for the shape.
* `available_core_count_per_node` - The maximum number of CPU cores per DB node that can be enabled for the shape.
* `available_data_storage_in_tbs` - The maximum amount of data storage, in terabytes (TB), that can be enabled for the shape.
* `available_data_storage_per_server_in_tbs` - The maximum amount of data storage, in terabytes (TB), that's available per storage server for the shape.
* `available_db_node_per_node_in_gbs` - The maximum amount of DB node storage, in gigabytes (GB), that's available per DB node for the shape.
* `available_db_node_storage_in_gbs` - The maximum amount of DB node storage, in gigabytes (GB), that can be enabled for the shape.
* `available_memory_in_gbs` - The maximum amount of memory, in gigabytes (GB), that can be enabled for the shape.
* `available_memory_per_node_in_gbs` - The maximum amount of memory, in gigabytes (GB), that's available per DB node for the shape.
* `compute_model` - The OCI compute model used when creating or cloning an instance: ECPU or OCPU.
* `core_count_increment` - The discrete number by which the CPU core count for the shape can be increased or decreased.
* `max_storage_count` - The maximum number of Exadata storage servers available for the shape.
* `maximum_node_count` - The maximum number of compute servers available for the shape.
* `min_core_count_per_node` - The minimum number of CPU cores that can be enabled per node for the shape.
* `min_data_storage_in_tbs` - The minimum amount of data storage, in terabytes (TB), that must be allocated for the shape.
* `min_db_node_storage_per_node_in_gbs` - The minimum amount of DB node storage, in gigabytes (GB), that must be allocated per DB node for the shape.
* `min_memory_per_node_in_gbs` - The minimum amount of memory, in gigabytes (GB), that must be allocated per DB node for the shape.
* `min_storage_count` - The minimum number of Exadata storage servers available for the shape.
* `minimum_core_count` - The minimum number of CPU cores that can be enabled for the shape.
* `minimum_node_count` - The minimum number of compute servers available for the shape.
* `name` - The name of the shape.
* `runtime_minimum_core_count` - The runtime minimum number of CPU cores that can be enabled for the shape.
* `shape_family` - The family of the shape.
* `shape_type` - The shape type, determined by the CPU hardware.
