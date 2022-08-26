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

The following arguments are supported:

* `base_capacity` - (Optional) The base data warehouse capacity of the workgroup in Redshift Processing Units (RPUs).
* `config_parameter` - (Optional) An array of parameters to set for more control over a serverless database. See `Config Parameter` below.
* `enhanced_vpc_routing` - (Optional) The value that specifies whether to turn on enhanced virtual private cloud (VPC) routing, which forces Amazon Redshift Serverless to route traffic through your VPC instead of over the internet.
* `publicly_accessible` - (Optional) A value that specifies whether the workgroup can be accessed from a public network.
* `security_group_ids` - (Optional) An array of security group IDs to associate with the workgroup.
* `subnet_ids` - (Optional) An array of VPC subnet IDs to associate with the workgroup.
* `security_group_ids` - (Optional) An array of security group IDs to associate with the workgroup.
* `workgroup_name` - (Required) The name of the workgroup.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Config Parameter

* `parameter_key` - (Required) The key of the parameter. The options are `datestyle`, `enable_user_activity_logging`, `query_group`, `search_path`, and `max_query_execution_time`.
* `parameter_value` - (Required) The value of the parameter to set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Redshift Serverless Workgroup.
* `id` - The Redshift Workgroup Name.
* `workgroup_id` - The Redshift Workgroup ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Redshift Serverless Workgroups can be imported using the `workgroup_name`, e.g.,

```
$ terraform import aws_redshiftserverless_workgroup.example example
```
