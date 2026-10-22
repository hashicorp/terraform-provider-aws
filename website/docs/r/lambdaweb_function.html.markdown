---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_function"
description: |-
  Manages an AWS Lambda Web function.
---

# Resource: aws_lambdaweb_function

Manages an AWS Lambda Web function.

~> **Note:** This resource creates the web function only, with no revisions and no endpoints. The web function serves no traffic until you add revisions and endpoints. To add them, use the `aws_lambdaweb_function_revision` and `aws_lambdaweb_function_endpoint` resources.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambdaweb_function" "example" {
  function_name = "example-function"
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Web function name.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the function.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambdaweb_function.example
  identity = {
    function_name = "example-function"
  }
}

resource "aws_lambdaweb_function" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `function_name` (String) Web function name.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a function using `function_name`. For example:

```terraform
import {
  to = aws_lambdaweb_function.example
  id = "example-function"
}
```

Using `terraform import`, import a function using `function_name`. For example:

```console
% terraform import aws_lambdaweb_function.example example-function
```
