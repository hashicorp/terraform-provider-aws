---
subcategory: "CloudWatch Evidently"
layout: "aws"
page_title: "AWS: aws_evidently_segment"
description: |-
  Provides a CloudWatch Evidently Segment resource.
---

# Resource: aws_evidently_segment

Provides a CloudWatch Evidently Segment resource.

## Example Usage

### Basic

```terraform
resource "aws_evidently_segment" "example" {
  name    = "example"
  pattern = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"

  tags = {
    "Key1" = "example Segment"
  }
}
```

### With JSON object in pattern

```terraform
resource "aws_evidently_segment" "example" {
  name    = "example"
  pattern = <<JSON
  {
    "Price": [
      {
        "numeric": [">",10,"<=",20]
      }
    ]
  }
  JSON

  tags = {
    "Key1" = "example Segment"
  }
}
```

### With Description

```terraform
resource "aws_evidently_segment" "example" {
  name        = "example"
  pattern     = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"
  description = "example"
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional, Forces new resource) Specifies the description of the segment.
* `name` - (Required, Forces new resource) A name for the segment.
* `pattern` - (Required, Forces new resource) The pattern to use for the segment. For more information about pattern syntax, see [Segment rule pattern syntax](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-Evidently-segments.html#CloudWatch-Evidently-segments-syntax.html).
* `tags` - (Optional) Tags to apply to the segment. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the segment.
* `created_time` - The date and time that the segment is created.
* `experiment_count` - The number of experiments that this segment is used in. This count includes all current experiments, not just those that are currently running.
* `id` - The ID has the same value as the ARN of the segment.
* `last_updated_time` - The date and time that this segment was most recently updated.
* `launch_count` - The number of launches that this segment is used in. This count includes all current launches, not just those that are currently running.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

CloudWatch Evidently Segment can be imported using the `arn`, e.g.,

```
$ terraform import aws_evidently_segment.example arn:aws:evidently:us-west-2:123456789012:segment/example
```
