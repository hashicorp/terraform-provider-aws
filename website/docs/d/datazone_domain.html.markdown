---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_domain"
description: |-
  Terraform data source for managing an AWS DataZone Domain.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->

# Data Source: aws_datazone_domain

Terraform data source for managing an AWS DataZone Domain.

## Example Usage

### Basic Usage

```terraform
data "aws_datazone_domain" "example" {
  name = "example_domain"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the DataZone Domain. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `id` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
