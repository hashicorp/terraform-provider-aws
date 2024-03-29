---
subcategory: "Data Pipeline"
layout: "aws"
page_title: "AWS: aws_datapipeline_pipeline_definition"
description: |-
  Provides a DataPipeline Definition.
---

# Resource: aws_datapipeline_pipeline_definition

Provides a DataPipeline Pipeline Definition resource.

## Example Usage

```terraform
resource "aws_datapipeline_pipeline" "default" {
  name = "tf-pipeline-default"
}

resource "aws_datapipeline_pipeline_definition" "example" {
  pipeline_id = aws_datapipeline_pipeline.default.id
  pipeline_object {
    id   = "Default"
    name = "Default"
    field {
      key          = "workerGroup"
      string_value = "workerGroup"
    }
  }
  pipeline_object {
    id   = "Schedule"
    name = "Schedule"
    field {
      key          = "startDateTime"
      string_value = "2012-12-12T00:00:00"
    }
    field {
      key          = "type"
      string_value = "Schedule"
    }
    field {
      key          = "period"
      string_value = "1 hour"
    }
    field {
      key          = "endDateTime"
      string_value = "2012-12-21T18:00:00"
    }
  }
  pipeline_object {
    id   = "SayHello"
    name = "SayHello"
    field {
      key          = "type"
      string_value = "ShellCommandActivity"
    }
    field {
      key          = "command"
      string_value = "echo hello"
    }
    field {
      key          = "parent"
      string_value = "Default"
    }
    field {
      key          = "schedule"
      string_value = "Schedule"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `pipeline_id` - (Required) ID of the pipeline.
* `pipeline_object` - (Required) Configuration block for the objects that define the pipeline. See below

The following arguments are optional:

* `parameter_object` - (Optional) Configuration block for the parameter objects used in the pipeline definition. See below
* `parameter_value` - (Optional) Configuration block for the parameter values used in the pipeline definition. See below

### `pipeline_object`

* `field` - (Required) Configuration block for Key-value pairs that define the properties of the object. See below
* `id` - (Required) ID of the object.
* `name` - (Required) ARN of the storage connector.

### `field`

* `key` - (Required) Field identifier.
* `ref_value` - (Optional) Field value, expressed as the identifier of another object
* `string_value` - (Optional) Field value, expressed as a String.

### `parameter_object`

* `attribute` - (Required) Configuration block for attributes of the parameter object. See below
* `id` - (Required) ID of the parameter object.

### `attribute`

* `key` - (Required) Field identifier.
* `string_value` - (Required) Field value, expressed as a String.

### `parameter_value`

* `id` - (Required) ID of the parameter value.
* `string_value` - (Required) Field value, expressed as a String.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique ID of the datapipeline definition.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_datapipeline_pipeline_definition` using the id. For example:

```terraform
import {
  to = aws_datapipeline_pipeline_definition.example
  id = "df-1234567890"
}
```

Using `terraform import`, import `aws_datapipeline_pipeline_definition` using the id. For example:

```console
% terraform import aws_datapipeline_pipeline_definition.example df-1234567890
```
