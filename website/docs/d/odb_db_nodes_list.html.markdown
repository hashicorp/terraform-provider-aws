---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_db_nodes_list"
description: |-
  Provides details about an AWS Oracle Database@AWS Db Nodes List.
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

# Data Source: aws_odb_db_nodes_list

Provides details about an AWS Oracle Database@AWS Db Nodes List.

## Example Usage

### Basic Usage

```terraform
data "aws_odb_db_nodes_list" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Db Nodes List.
* `example_attribute` - Brief description of the attribute.
