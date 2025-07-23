---
subcategory: "EventBridge"
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
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name_prefix` - (Optional) Specifying this limits the results to only those partner event sources with names that start with the specified prefix

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the partner event source
* `created_by` - Name of the SaaS partner that created the event source
* `name` - Name of the event source
* `state` - State of the event source (`ACTIVE` or `PENDING`)
