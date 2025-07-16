---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_networks_list"
description: |-
  Provides details about an AWS Oracle Database@AWS Networks List.
---
<!---
Documentation guidelines:
- Begin data source descriptions with "Provides details about..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Data Source: aws_odb_networks_list

Provides details about an AWS Oracle Database@AWS Networks List.

## Example Usage

### Basic Usage

```terraform
data "aws_odb_networks_list" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Networks List.
* `example_attribute` - Brief description of the attribute.
