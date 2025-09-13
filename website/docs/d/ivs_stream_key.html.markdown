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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `channel_arn` - (Required) ARN of the Channel.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Stream Key.
* `tags` - Map of tags assigned to the resource.
* `value` - Stream Key value.
