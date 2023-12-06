---
subcategory: "Data Pipeline"
layout: "aws"
page_title: "AWS: aws_datapipeline_pipeline_definition"
description: |-
  Provides details about a specific DataPipeline Definition.
---

# Source: aws_datapipeline_pipeline_definition

Provides details about a specific DataPipeline Pipeline Definition.

## Example Usage

```terraform
data "aws_datapipeline_pipeline_definition" "example" {
  pipeline_id = "pipelineID"
}
```

## Argument Reference

The following arguments are required:

* `pipeline_id` - (Required) ID of the pipeline.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `parameter_object` - Parameter objects used in the pipeline definition. See below
* `parameter_value` - Parameter values used in the pipeline definition. See below
* `pipeline_object` - Objects defined in the pipeline. See below

### `parameter_object`

* `attributes` - Attributes of the parameter object. See below
* `id` - ID of the parameter object.

### `attributes`

* `key` - Field identifier.
* `string_value` - Field value, expressed as a String.

### `parameter_value`

* `id` - ID of the parameter value.
* `string_value` - Field value, expressed as a String.

### `pipeline_object`

* `field` - Key-value pairs that define the properties of the object. See below
* `id` - ID of the object.
* `name` - ARN of the storage connector.

### `field`

* `key` - Field identifier.
* `ref_value` - Field value, expressed as the identifier of another object
* `string_value` - Field value, expressed as a String.
