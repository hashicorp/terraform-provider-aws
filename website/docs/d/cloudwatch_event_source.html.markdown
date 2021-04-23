---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_source"
description: |-
  Get information on an EventBridge (Cloudwatch) Event Source.
---

# Data Source: aws_cloudwatch_event_source

Use this data source to get information about an EventBridge Partner Event Source. This data source will only return one partner event source. An error will be returned if multiple sources match the same name prefix.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
data "aws_cloudwatch_event_source" "examplepartner" {
  name_prefix = "aws.partner/examplepartner.com"
}

resource "aws_cloudwatch_event_bus" "examplepartner" {
  name              = data.aws_cloudwatch_event_source.examplepartner.name
  event_source_name = data.aws_cloudwatch_event_source.examplepartner.name
}
```

## Argument Reference

The following arguments are supported:

* `name_prefix` - (Optional) A name prefix to filter results returned. Only API destinations with a name that starts with the prefix are returned

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Partner event source
* `created_by` - The name of the SaaS partner that created the event source.
* `state` - The state of the event source (`ACTIVE` or `PENDING`)
