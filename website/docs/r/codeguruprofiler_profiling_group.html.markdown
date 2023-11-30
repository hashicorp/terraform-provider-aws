---
subcategory: "CodeGuru Profiler"
layout: "aws"
page_title: "AWS: aws_codeguruprofiler_profiling_group"
description: |-
  Terraform resource for managing an AWS CodeGuru Profiler Profiling Group.
---
# Resource: aws_codeguruprofiler_profiling_group

Terraform resource for managing an AWS CodeGuru Profiler Profiling Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_codeguruprofiler_profiling_group" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `tags` - (Optional) A map of tags assigned to the WorkSpaces Connection Alias. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Profiling Group. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeGuru Profiler Profiling Group using the `example_id_arg`. For example:

```terraform
import {
  to = aws_codeguruprofiler_profiling_group.example
  id = "profiling_group-id-12345678"
}
```

Using `terraform import`, import CodeGuru Profiler Profiling Group using the `example_id_arg`. For example:

```console
% terraform import aws_codeguruprofiler_profiling_group.example profiling_group-id-12345678
```
