---
subcategory: "Data Pipeline"
layout: "aws"
page_title: "AWS: aws_datapipeline_pipeline"
description: |-
  Provides details about a specific DataPipeline.
---

# Source: aws_datapipeline_pipeline

Provides details about a specific DataPipeline Pipeline.

## Example Usage

```terraform
data "aws_datapipeline_pipeline" "example" {
  pipeline_id = "pipelineID"
}
```

## Argument Reference

The following arguments are required:

* `pipeline_id` - (Required) ID of the pipeline.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `name` - Name of Pipeline.
* `description` - Description of Pipeline.
* `tags` - Map of tags assigned to the resource.
