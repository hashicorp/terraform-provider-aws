---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_workgroup"
description: |-
  Provides a Redshift Serverless Workgroup resource.
---

# Resource: aws_redshiftserverless_workgroup

Creates a new Amazon Redshift Serverless Workgroup.

## Example Usage

```terraform
resource "aws_redshiftserverless_workgroup" "example" {
  namespace_name = "concurrency-scaling"
  workgroup_name = "concurrency-scaling"
}
```

## Argument Reference

The following arguments are required:

* `namespace_name` - (Required) The name of the namespace.
* `workgroup_name` - (Required) The name of the workgroup.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `base_capacity` - (Optional) The base data warehouse capacity of the workgroup in Redshift Processing Units (RPUs).
* `price_performance_target` - (Optional) Price-performance scaling for the workgroup. See `Price Performance Target` below.
* `config_parameter` - (Optional) An array of parameters to set for more control over a serverless database. See `Config Parameter` below.
* `enhanced_vpc_routing` - (Optional) The value that specifies whether to turn on enhanced virtual private cloud (VPC) routing, which forces Amazon Redshift Serverless to route traffic through your VPC instead of over the internet.
* `max_capacity` - (Optional) The maximum data-warehouse capacity Amazon Redshift Serverless uses to serve queries, specified in Redshift Processing Units (RPUs).
* `port` - (Optional) The port number on which the cluster accepts incoming connections.
* `publicly_accessible` - (Optional) A value that specifies whether the workgroup can be accessed from a public network.
* `security_group_ids` - (Optional) An array of security group IDs to associate with the workgroup.
* `subnet_ids` - (Optional) An array of VPC subnet IDs to associate with the workgroup. When set, must contain at least three subnets spanning three Availability Zones. A minimum number of IP addresses is required and scales with the Base Capacity. For more information, see the following [AWS document](https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-known-issues.html).
* `track_name` - (Optional) The name of the track for the workgroup. If it is `current`, you get the most up-to-date certified release version with the latest features, security updates, and performance enhancements. If it is `trailing`, you will be on the previous certified release. For more information, see the following [AWS document](https://docs.aws.amazon.com/redshift/latest/mgmt/tracks.html).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Price Performance Target

* `enabled` - (Required) Whether to enable price-performance scaling.
* `level` - (Required) The price-performance scaling level. Valid values are `1` (LOW_COST), `25` (ECONOMICAL), `50` (BALANCED), `75` (RESOURCEFUL), and `100` (HIGH_PERFORMANCE).

### Config Parameter

* `parameter_key` - (Required) The key of the parameter. The options are `auto_mv`, `datestyle`, `enable_case_sensitive_identifier`, `enable_user_activity_logging`, `query_group`, `search_path`, `require_ssl`, `use_fips_ssl`, and [query monitoring metrics](https://docs.aws.amazon.com/redshift/latest/dg/cm-c-wlm-query-monitoring-rules.html#cm-c-wlm-query-monitoring-metrics-serverless) that let you define performance boundaries: `max_query_cpu_time`, `max_query_blocks_read`, `max_scan_row_count`, `max_query_execution_time`, `max_query_queue_time`, `max_query_cpu_usage_percent`, `max_query_temp_blocks_to_disk`, `max_join_row_count` and `max_nested_loop_join_row_count`.
* `parameter_value` - (Required) The value of the parameter to set.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Redshift Serverless Workgroup.
* `id` - The Redshift Workgroup Name.
* `workgroup_id` - The Redshift Workgroup ID.
* `endpoint` - The endpoint that is created from the workgroup. See `Endpoint` below.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### Endpoint

* `address` - The DNS address of the VPC endpoint.
* `port` - The port that Amazon Redshift Serverless listens on.
* `vpc_endpoint` - The VPC endpoint or the Redshift Serverless workgroup. See `VPC Endpoint` below.

#### VPC Endpoint

* `vpc_endpoint_id` - The DNS address of the VPC endpoint.
* `vpc_id` - The port that Amazon Redshift Serverless listens on.
* `network_interface` - The network interfaces of the endpoint.. See `Network Interface` below.

##### Network Interface

* `availability_zone` - The availability Zone.
* `network_interface_id` - The unique identifier of the network interface.
* `private_ip_address` - The IPv4 address of the network interface within the subnet.
* `subnet_id` - The unique identifier of the subnet.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `20m`)
- `update` - (Default `20m`)
- `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Serverless Workgroups using the `workgroup_name`. For example:

```terraform
import {
  to = aws_redshiftserverless_workgroup.example
  id = "example"
}
```

Using `terraform import`, import Redshift Serverless Workgroups using the `workgroup_name`. For example:

```console
% terraform import aws_redshiftserverless_workgroup.example example
```
