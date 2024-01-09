---
subcategory: "Compute Optimizer"
layout: "aws"
page_title: "AWS: aws_computeoptimizer_recommendation_preferences"
description: |-
  Terraform resource for managing an AWS Compute Optimizer Recommendation Preferences.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_computeoptimizer_recommendation_preferences

Terraform resource for managing an AWS Compute Optimizer Recommendation Preferences.

## Example Usage

### Basic Usage

```terraform
resource "aws_computeoptimizer_recommendation_preferences" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Recommendation Preferences. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Compute Optimizer Recommendation Preferences using the `example_id_arg`. For example:

```terraform
import {
  to = aws_computeoptimizer_recommendation_preferences.example
  id = "recommendation_preferences-id-12345678"
}
```

Using `terraform import`, import Compute Optimizer Recommendation Preferences using the `example_id_arg`. For example:

```console
% terraform import aws_computeoptimizer_recommendation_preferences.example recommendation_preferences-id-12345678
```
