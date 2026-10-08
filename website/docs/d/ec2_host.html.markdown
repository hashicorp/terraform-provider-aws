---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_host"
description: |-
  Get information on an EC2 Host.
---

# Data Source: aws_ec2_host

Use this data source to get information about an EC2 Dedicated Host.

## Example Usage

```terraform
resource "aws_ec2_host" "test" {
  instance_type     = "c5.18xlarge"
  availability_zone = "us-west-2a"
}

data "aws_ec2_host" "test" {
  host_id = aws_ec2_host.test.id
}
```

### Filter Example

```terraform
data "aws_ec2_host" "test" {
  filter {
    name   = "instance-type"
    values = ["c5.18xlarge"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Configuration block. Detailed below.
* `host_id` - (Optional) ID of the Dedicated Host.

The arguments of this data source act as filters for querying the available EC2 Hosts in the current region.
The given filters must match exactly one host whose data will be exported as attributes.

### filter

This block allows for complex filters. You can use one or more `filter` blocks.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeHosts.html).
* `values` - (Required) Set of values that are accepted for the given field. A host will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the Dedicated Host.
* `allocation_time` - Time that the Dedicated Host was allocated, in RFC3339 format.
* `allows_multiple_instance_types` - Whether the Dedicated Host supports multiple instance types of the same instance family. Valid values: `on`, `off`.
* `arn` - ARN of the Dedicated Host.
* `asset_id` - The ID of the Outpost hardware asset on which the Dedicated Host is allocated.
* `auto_placement` - Whether auto-placement is on or off.
* `available_capacity` - Information about the available capacity of the Dedicated Host. See [`available_capacity`](#available_capacity) below.
* `availability_zone` - Availability Zone of the Dedicated Host.
* `availability_zone_id` - AZ ID of the Availability Zone in which the Dedicated Host is allocated (e.g., `use1-az1`).
* `cores` - Number of cores on the Dedicated Host.
* `host_maintenance` - Whether host maintenance is enabled or disabled for the Dedicated Host. Valid values: `on`, `off`.
* `host_recovery` - Whether host recovery is enabled or disabled for the Dedicated Host.
* `host_reservation_id` - The reservation ID of the Dedicated Host.
* `instance_family` - Instance family supported by the Dedicated Host. For example, "m5".
* `instance_type` - Instance type supported by the Dedicated Host. For example, "m5.large". If the host supports multiple instance types, no instanceType is returned.
* `instances` - The instances running on the Dedicated Host. See [`instances`](#instances) below.
* `member_of_service_linked_resource_group` - Whether the Dedicated Host is in a host resource group.
* `outpost_arn` - ARN of the AWS Outpost on which the Dedicated Host is allocated.
* `owner_id` - ID of the AWS account that owns the Dedicated Host.
* `release_time` - Time that the Dedicated Host was released, in RFC3339 format.
* `sockets` - Number of sockets on the Dedicated Host.
* `state` - Allocation state of the Dedicated Host. Valid values: `available`, `under-assessment`, `permanent-failure`, `released`, `released-permanent-failure`, `pending`.
* `total_vcpus` - Total number of vCPUs on the Dedicated Host.

### `available_capacity`

* `available_instance_capacity` - The number of instances that can be launched onto the Dedicated Host for each instance size supported. See [`available_instance_capacity`](#available_instance_capacity) below.
* `available_vcpus` - The number of vCPUs available for launching instances onto the Dedicated Host.

### `available_instance_capacity`

* `available_capacity` - The number of instances that can be launched onto the Dedicated Host based on the host's available capacity.
* `instance_type` - The instance type supported by the Dedicated Host.
* `total_capacity` - The total number of instances that can be launched onto the Dedicated Host if there are no instances running on it.

### `instances`

* `instance_id` - The ID of the instance running on the Dedicated Host.
* `instance_type` - The instance type of the running instance.
* `owner_id` - The ID of the AWS account that owns the instance.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
