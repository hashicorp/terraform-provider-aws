---
subcategory: "CloudWatch Evidently"
layout: "aws"
page_title: "AWS: aws_evidently_project"
description: |-
  Provides a CloudWatch Evidently Project resource.
---

# Resource: aws_evidently_project

Provides a CloudWatch Evidently Project resource.

## Example Usage

### Basic

```terraform
resource "aws_evidently_project" "example" {
  name        = "Example"
  description = "Example Description"

  tags = {
    "Key1" = "example Project"
  }
}
```

### Store evaluation events in a CloudWatch Log Group

```terraform
resource "aws_evidently_project" "example" {
  name        = "Example"
  description = "Example Description"

  data_delivery {
    cloudwatch_logs {
      log_group = "example-log-group-name"
    }
  }

  tags = {
    "Key1" = "example Project"
  }
}
```

### Store evaluation events in an S3 bucket

```terraform
resource "aws_evidently_project" "example" {
  name        = "Example"
  description = "Example Description"

  data_delivery {
    s3_destination {
      bucket = "example-bucket-name"
      prefix = "example"
    }
  }

  tags = {
    "Key1" = "example Project"
  }
}
```

## Argument Reference

The following arguments are supported:

* `data_delivery` - (Optional) A block that contains information about where Evidently is to store evaluation events for longer term storage, if you choose to do so. If you choose not to store these events, Evidently deletes them after using them to produce metrics and other experiment results that you can view. See below.
* `description` - (Optional) Specifies the description of the project.
* `name` - (Required) A name for the project.
* `tags` - (Optional) Tags to apply to the project. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `data_delivery` block supports the following arguments:

~> **NOTE:** You can't specify both `cloudwatch_logs` and `s3_destination`.

* `cloudwatch_logs` - (Optional) A block that defines the CloudWatch Log Group that stores the evaluation events. See below.
* `s3_destination` - (Optional) A block that defines the S3 bucket and prefix that stores the evaluation events. See below.

The `cloudwatch_logs` block supports the following arguments:

* `log_group` - (Optional) The name of the log group where the project stores evaluation events.

The `s3_destination` block supports the following arguments:

* `bucket` - (Optional) The name of the bucket in which Evidently stores evaluation events.
* `prefix` - (Optional) The bucket prefix in which Evidently stores evaluation events.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `2m`)
* `delete` - (Default `2m`)
* `update` - (Default `2m`)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `active_experiment_count` - The number of ongoing experiments currently in the project.
* `active_launch_count` - The number of ongoing launches currently in the project.
* `arn` - The ARN of the project.
* `created_time` - The date and time that the project is created.
* `experiment_count` - The number of experiments currently in the project. This includes all experiments that have been created and not deleted, whether they are ongoing or not.
* `feature_count` - The number of features currently in the project.
* `id` - The ID has the same value as the arn of the project.
* `last_updated_time` - The date and time that the project was most recently updated.
* `launch_count` - The number of launches currently in the project. This includes all launches that have been created and not deleted, whether they are ongoing or not.
* `status` - The current state of the project. Valid values are `AVAILABLE` and `UPDATING`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

CloudWatch Evidently Project can be imported using the `arn`, e.g.,

```
$ terraform import aws_evidently_project.example arn:aws:evidently:us-east-1:123456789012:segment/example
```
