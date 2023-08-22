---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_aggregate_authorization"
description: |-
  Manages an AWS Config Aggregate Authorization.
---

# Resource: aws_config_aggregate_authorization

Manages an AWS Config Aggregate Authorization

## Example Usage

```terraform
resource "aws_config_aggregate_authorization" "example" {
  account_id = "123456789012"
  region     = "eu-west-2"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Required) Account ID
* `region` - (Required) Region
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the authorization
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Config aggregate authorizations using `account_id:region`. For example:

```terraform
import {
  to = aws_config_aggregate_authorization.example
  id = "123456789012:us-east-1"
}
```

Using `terraform import`, import Config aggregate authorizations using `account_id:region`. For example:

```console
% terraform import aws_config_aggregate_authorization.example 123456789012:us-east-1
```
