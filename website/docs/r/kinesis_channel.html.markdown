---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_channel"
description: |-
  Manages an AWS Kinesis Channel.
---
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Don't spell out common abbreviations (e.g. use "ARN", not "Amazon Resource Name")
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_kinesis_channel

Manages an AWS Kinesis Channel.

## Example Usage

### Basic Usage

```terraform
resource "aws_kinesis_channel" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Channel.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_kinesis_channel.example
  identity = {
<!---
Add only required attributes in this example.
--->
  }
}

resource "aws_kinesis_channel" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required
<!---
Required attributes here:
> ARN Identity:
* `arn` - ARN of the Channel.
> Parameterized Identity:
* `example_id_arg` - ID argument of the Channel.
> Singleton Identity: no required attributes.
--->

#### Optional
<!---
Optional attributes here:
> ARN Identity: no optional attributes.
> Parameterized Identity and Singleton Identity: remove `region` if the resource is global.
--->
* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Kinesis Channel using the `example_id_arg`. For example:

```terraform
import {
  to = aws_kinesis_channel.example
  id = "channel-id-12345678"
}
```

Using `terraform import`, import Kinesis Channel using the `example_id_arg`. For example:

```console
% terraform import aws_kinesis_channel.example channel-id-12345678
```
