---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_cloud_exadata_infrastructure"
description: |-
  Manages an AWS Oracle Database@AWS Cloud Exadata Infrastructure.
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

# Resource: aws_odb_cloud_exadata_infrastructure

Manages an AWS Oracle Database@AWS Cloud Exadata Infrastructure.

## Example Usage

### Basic Usage

```terraform
resource "aws_odb_cloud_exadata_infrastructure" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Cloud Exadata Infrastructure.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Oracle Database@AWS Cloud Exadata Infrastructure using the `example_id_arg`. For example:

```terraform
import {
  to = aws_odb_cloud_exadata_infrastructure.example
  id = "cloud_exadata_infrastructure-id-12345678"
}
```

Using `terraform import`, import Oracle Database@AWS Cloud Exadata Infrastructure using the `example_id_arg`. For example:

```console
% terraform import aws_odb_cloud_exadata_infrastructure.example cloud_exadata_infrastructure-id-12345678
```
