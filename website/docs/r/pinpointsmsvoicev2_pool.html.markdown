---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_pool"
description: |-
  Manages an AWS End User Messaging SMS Pool.
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

# Resource: aws_pinpointsmsvoicev2_pool

Manages an AWS End User Messaging SMS Pool.

## Example Usage

### Basic Usage

```terraform
resource "aws_pinpointsmsvoicev2_pool" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Pool.
* `example_attribute` - Brief description of the attribute.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_pool.example
  identity = {
<!---
Add only required attributes in this example.
--->
  }
}

resource "aws_pinpointsmsvoicev2_pool" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required
<!---
Required attributes here:
> ARN Identity:
* `arn` - ARN of the Pool.
> Parameterized Identity:
* `example_id_arg` - ID argument of the Pool.
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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import End User Messaging SMS Pool using the `example_id_arg`. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_pool.example
  id = "pool-id-12345678"
}
```

Using `terraform import`, import End User Messaging SMS Pool using the `example_id_arg`. For example:

```console
% terraform import aws_pinpointsmsvoicev2_pool.example pool-id-12345678
```
