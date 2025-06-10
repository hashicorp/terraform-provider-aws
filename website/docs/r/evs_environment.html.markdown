---
subcategory: "EVS"
layout: "aws"
page_title: "AWS: aws_evs_environment"
description: |-
  Manages an Amazon EVS Environment.
---

# Resource: aws_evs_environment

Manages an Amazon EVS Environment. Use this resource to create an environment that runs VCF software.

## Example Usage

```terraform
resource "aws_evs_environment" "example" {}
```

## Argument Reference

The following arguments are required:

The following arguments are optional:

* `tags` - (Optional) Key-value map of tags for the environment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `environment_arn` - Environment ARN.
* `environment_id` - Environment ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `45m`)
* `delete` - (Default `45m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import environments using the `environment_id`. For example:

```terraform
import {
  to = aws_evs_environment.example
  id = "env-abcde12345"
}
```

Using `terraform import`, import environments using the `environment_id`. For example:

```console
% terraform import aws_evs_environment.example env-abcde12345
```
