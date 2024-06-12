---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_registry"
description: |-
  Terraform data source for managing an AWS Glue Registry.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->

# Data Source: aws_glue_registry

Terraform data source for managing an AWS Glue Registry.

## Example Usage

### Basic Usage

```terraform
data "aws_glue_registry" "example" {
  id = "arn:aws:glue:us-west-2:123456789012:registry/example"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Amazon Resource Name (ARN) of Glue Registry.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of Glue Registry.
* `description` - A description of the registry.
* `registry_name` - The Name of the registry.
