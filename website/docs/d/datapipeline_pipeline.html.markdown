---
subcategory: "Data Pipeline"
layout: "aws"
page_title: "AWS: aws_datapipeline_pipeline"
description: |-
  Provides details about a specific DataPipeline.
---

# Data Source: aws_datapipeline_pipeline

Provides details about a specific DataPipeline Pipeline.

## Example Usage

```terraform
data "aws_datapipeline_pipeline" "example" {
  pipeline_id = "pipelineID"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `pipeline_id` - (Required) ID of the pipeline.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `name` - Name of Pipeline.
* `description` - Description of Pipeline.
* `tags` - Map of tags assigned to the resource.
