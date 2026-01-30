---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_parameter_version_labels"
description: |-
  Manages an AWS SSM (Systems Manager) Parameter Version Labels.
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

# Resource: aws_ssm_parameter_version_labels

Manages an AWS SSM (Systems Manager) Parameter Version Labels.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssm_parameter_version_labels" "example" {
  name   = "parameter"
  labels = ["label1", "label2"]
}
```

## Argument Reference

The following arguments are required:

- `name` - (Required) SSM Parameter name.
- `labels` - (Required) a list of labels to add

The following arguments are optional:

- `version` - (Optional) SSM Parameter version. If omitted, latest parameter version is used

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM (Systems Manager) Parameter Version Labels using the `name` or `name:version`. For example:

```terraform
import {
  to = aws_ssm_parameter_version_labels.example
  id = "parameter"
}
```

```terraform
import {
  to = aws_ssm_parameter_version_labels.example
  id = "parameter:1"
}
```

Using `terraform import`, import SSM (Systems Manager) Parameter Version Labels using the `name` or `name:version`. For example:

```console
% terraform import aws_ssm_parameter_version_labels.example parameter
```

```console
% terraform import aws_ssm_parameter_version_labels.example parameter:1
```
