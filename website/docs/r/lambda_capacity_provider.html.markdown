---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_capacity_provider"
description: |-
  Manages an AWS Lambda Capacity Provider.
---

# Resource: aws_lambda_capacity_provider

Manages an AWS Lambda Capacity Provider.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambda_capacity_provider" "example" {
  name = "example"

  vpc_config {
    subnet_ids         = aws_subnet.example[*].id
    security_group_ids = [aws_security_group.example.id]
  }

  permissions_config {
    capacity_provider_operator_role_arn = aws_iam_role.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the Capacity Provider.
* `vpc_config` - (Required) Configuration block for VPC settings. See [VPC Config](#vpc_config) below.
* `permissions_config` - (Required) Configuration block for permissions settings. See [Permissions Config](#permissions_config) below.

The following arguments are optional:

* `capacity_provider_scaling_policy` - (Optional) Configuration block for scaling policy settings. See [Capacity Provider Scaling Policy](#capacity_provider_scaling_policy) below.
* `instance_requirements` - (Optional) Configuration block for instance requirements settings. See [Instance Requirements](#instance_requirements) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### vpc_config

* `subnet_ids` - (Required) List of subnet IDs for the VPC.
* `security_group_ids` - (Required) List of security group IDs for the VPC.

### permissions_config

* `capacity_provider_operator_role_arn` - (Required) The ARN of the IAM role that allows Lambda to manage the Capacity Provider.

### capacity_provider_scaling_policy

* `max_vpcu_count` - (Optional) The maximum number of VPCUs for the Capacity Provider.
* `scaling_mode` - (Required) The scaling mode for the Capacity Provider. Valid values are `AUTO` and `MANUAL`. Defaults to `AUTO`.
* `scaling_policies` - (Optional) List of scaling policies. See [Scaling Policies](#scaling_policies) below.

#### scaling_policies

* `predefined_metric_type` - (Required) The predefined metric type for the scaling policy. Valid values are `LAMBDA_PROVISIONED_CONCURRENCY_UTILIZATION`.
* `target_value` - (Required) The target value for the scaling policy.

### instance_requirements

* `architectures` - (Required) List of CPU architectures. Valid values are `X86_64` and `ARM64`.
* `allowed_instance_types` - (Optional) List of allowed instance types.
* `excluded_instance_types` - (Optional) List of excluded instance types.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Capacity Provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Capacity Provider using the `name`. For example:

```terraform
import {
  to = aws_lambda_capacity_provider.example
  id = "example"
}
```

Using `terraform import`, import Lambda Capacity Provider using the `name`. For example:

```console
% terraform import aws_lambda_capacity_provider.example example
```
