---
subcategory: "CloudWatch Observability Access Manager"
layout: "aws"
page_title: "AWS: aws_oam_sink"
description: |-
  Terraform data source for managing an AWS CloudWatch Observability Access Manager Sink.
---

# Data Source: aws_oam_sink

Terraform data source for managing an AWS CloudWatch Observability Access Manager Sink.

## Example Usage

### Basic Usage

```terraform
data "aws_oam_sink" "example" {
  sink_identifier = "arn:aws:oam:us-west-1:111111111111:sink/abcd1234-a123-456a-a12b-a123b456c789"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sink_identifier` - (Required) ARN of the sink.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the sink.
* `id` - ARN of the sink. Use `arn` instead.
* `name` - Name of the sink.
* `sink_id` - Random ID string that AWS generated as part of the sink ARN.
* `tags` - Tags assigned to the sink.
