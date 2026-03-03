---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_associate_disassociate_iam_role"
description: |-
  Manages an AWS Oracle Database@AWS Associate Disassociate IAM Role.
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

# Resource: aws_odb_associate_disassociate_iam_role

Manages an AWS Oracle Database@AWS Associate Disassociate IAM Role.

## Example Usage

### Basic Usage

```terraform
resource "aws_odb_associate_disassociate_iam_role" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Associate Disassociate IAM Role.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Oracle Database@AWS Associate Disassociate IAM Role using the `example_id_arg`. For example:

```terraform
import {
  to = aws_odb_associate_disassociate_iam_role.example
  id = "associate_disassociate_iam_role-id-12345678"
}
```

Using `terraform import`, import Oracle Database@AWS Associate Disassociate IAM Role using the `example_id_arg`. For example:

```console
% terraform import aws_odb_associate_disassociate_iam_role.example associate_disassociate_iam_role-id-12345678
```
