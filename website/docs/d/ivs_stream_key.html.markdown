---
subcategory: "IVS (Interactive Video)"
layout: "aws"
page_title: "AWS: aws_ivs_stream_key"
description: |-
  Terraform data source for managing an AWS IVS (Interactive Video) Stream Key.
---

# Data Source: aws_ivs_stream_key

Terraform data source for managing an AWS IVS (Interactive Video) Stream Key.

## Example Usage

### Basic Usage

```terraform
data "aws_ivs_stream_key" "example" {
  channel_arn = "arn:aws:ivs:us-west-2:326937407773:channel/0Y1lcs4U7jk5"
}
```

## Argument Reference

The following arguments are required:

* `channel_arn` - (Required) ARN of the Channel.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Stream Key.
* `tags` - Map of tags assigned to the resource.
* `value` - Stream Key value.
