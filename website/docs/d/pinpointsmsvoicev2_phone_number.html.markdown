---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_phone_number"
description: |-
  Provides details about an AWS End User Messaging SMS Phone Number.
---
<!---
Documentation guidelines:
- Begin data source descriptions with "Provides details about..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Don't spell out common abbreviations (e.g. use "ARN", not "Amazon Resource Name")
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Data Source: aws_pinpointsmsvoicev2_phone_number

Provides details about an AWS End User Messaging SMS Phone Number.

## Example Usage

### Basic Usage

```terraform
data "aws_pinpointsmsvoicev2_phone_number" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Phone Number.
* `example_attribute` - Brief description of the attribute.
* `tags` - Map of tags assigned to the resource.
