---
subcategory: "CodeGuru Profiler"
layout: "aws"
page_title: "AWS: aws_codeguruprofiler_profiling_group"
description: |-
  Terraform data source for managing an AWS CodeGuru Profiler Profiling Group.
---

# Data Source: aws_codeguruprofiler_profiling_group

Terraform data source for managing an AWS CodeGuru Profiler Profiling Group.

## Example Usage

### Basic Usage

```terraform
data "aws_codeguruprofiler_profiling_group" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the profiling group.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `agent_orchestration_config` - Profiling Group agent orchestration config
* `arn` - ARN of the Profiling Group.
* `created_at`- Timestamp when Profiling Group was created.
* `compute_platform` - The compute platform of the profiling group.
* `profiling_status` - The status of the Profiling Group.
* `tags` - Mapping of Key-Value tags for the resource.
* `updated_at` -  Timestamp when Profiling Group was updated.
