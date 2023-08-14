---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_user"
description: |-
  Terraform resource for managing an AWS FinSpace Kx User.
---

# Resource: aws_finspace_kx_user

Terraform resource for managing an AWS FinSpace Kx User.

## Example Usage

### Basic Usage

```terraform
resource "aws_kms_key" "example" {
  description             = "Example KMS Key"
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "example" {
  name       = "my-tf-kx-environment"
  kms_key_id = aws_kms_key.example.arn
}

resource "aws_iam_role" "example" {
  name = "example-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_finspace_kx_user" "example" {
  name           = "my-tf-kx-user"
  environment_id = aws_finspace_kx_environment.example.id
  iam_role       = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) A unique identifier for the user.
* `environment_id` - (Required) Unique identifier for the KX environment.
* `iam_role` - (Required) IAM role ARN to be associated with the user.

The following arguments are optional:

* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX user.
* `id` - A comma-delimited string joining environment ID and user name.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS FinSpace Kx User using the `id` (environment ID and user name, comma-delimited). For example:

```terraform
import {
  to = aws_finspace_kx_user.example
  id = "n3ceo7wqxoxcti5tujqwzs,my-tf-kx-user"
}
```

Using `terraform import`, import an AWS FinSpace Kx User using the `id` (environment ID and user name, comma-delimited). For example:

```console
% terraform import aws_finspace_kx_user.example n3ceo7wqxoxcti5tujqwzs,my-tf-kx-user
```
