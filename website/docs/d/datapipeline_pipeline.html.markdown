---
subcategory: "DataPipeline"
layout: "aws"
page_title: "AWS: aws_datapipeline_pipeline"
sidebar_current: "docs-aws-resource-datapipeline-definition"
description: |-
  Provides details about a specific DataPipeline.
---

# Source: aws_datapipeline_pipeline

Provides details about a specific DataPipeline.

## Example Usage

```terraform
data "aws_datapipeline_pipeline" "example" {
  pipeline_id = "PIPELINE ID"
}
```

## Argument Reference

The following arguments are required:

* `pipeline_id` - (Required) ID of the pipeline.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `name` - Name of Pipeline.
* `description` - Description of Pipeline.
* `tags` - A map of tags to assign to the resource.

