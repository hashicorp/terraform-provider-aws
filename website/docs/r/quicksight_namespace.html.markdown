---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_namespace"
description: |-
  Terraform resource for managing an AWS QuickSight Namespace.
---

# Resource: aws_quicksight_namespace

Terraform resource for managing an AWS QuickSight Namespace.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_namespace" "example" {
  namespace = "example"
}
```

## Argument Reference

The following arguments are required:

* `namespace` - (Required) Name of the namespace.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID.
* `identity_store` - (Optional) User identity directory type. Defaults to `QUICKSIGHT`, the only current valid value.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Namespace.
* `capacity_region` - Namespace AWS Region.
* `creation_status` - Creation status of the namespace.
* `id` - A comma-delimited string joining AWS account ID and namespace.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `2m`)
* `delete` - (Default `2m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight Namespace using the AWS account ID and namespace separated by commas (`,`). For example:

```terraform
import {
  to = aws_quicksight_namespace.example
  id = "123456789012,example"
}
```

Using `terraform import`, import QuickSight Namespace using the AWS account ID and namespace separated by commas (`,`). For example:

```console
% terraform import aws_quicksight_namespace.example 123456789012,example
```
