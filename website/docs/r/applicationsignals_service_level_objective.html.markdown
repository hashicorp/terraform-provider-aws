---
subcategory: "Application Signals"
layout: "aws"
page_title: "AWS: aws_applicationsignals_service_level_objective"
description: |-
  Manages an AWS Application Signals Service Level Objective.
---
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_applicationsignals_service_level_objective

Manages an AWS Application Signals Service Level Objective.

## Example Usage

### Basic Usage

```terraform
resource "aws_applicationsignals_service_level_objective" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Level Objective.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Application Signals Service Level Objective using the `example_id_arg`. For example:

```terraform
import {
  to = aws_applicationsignals_service_level_objective.example
  id = "service_level_objective-id-12345678"
}
```

Using `terraform import`, import Application Signals Service Level Objective using the `example_id_arg`. For example:

```console
% terraform import aws_applicationsignals_service_level_objective.example service_level_objective-id-12345678
```
