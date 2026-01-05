---
subcategory: "secretsmanager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_tag"
description: |-
  Manages an individual AWS Secrets Manager secret tag
---

# Resource: aws_secretsmanager_tag

Manages an individual AWS Secrets Manager secret tag. This resource should only be used in cases where AWS Secrets Manager secrets are created outside Terraform (e.g., [AWS Secrets Manager secrets managed by other AWS services](https://docs.aws.amazon.com/secretsmanager/latest/userguide/service-linked-secrets.html), such as RDS).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_secretsmanager_secret` and `aws_secretsmanager_tag` to manage tags of the same AWS Secrets Manager secret will cause a perpetual difference where the `aws_secretsmanager_secret` resource will try to remove the tag being added by the `aws_secretsmanager_tag` resource. However, if the parent resource is created in the same configuration (i.e., if you have no other choice), you should add `ignore_changes = [tags]` in the parent resource's lifecycle block. This ensures that Terraform ignores differences in tags managed via the separate tagging resource, avoiding the perpetual difference mentioned above.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
resource "aws_secretsmanager_secret" "test" {
  name = "example-secret"

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_secretsmanager_tag" "test" {
  secret_id = aws_secretsmanager_secret.test.id
  key       = "ExampleKey"
  value     = "ExampleValue"
}
```

## Argument Reference

This resource supports the following arguments:

* `secret_id` - (Required) ID of the AWS Secrets Manager secret to tag.
* `key` - (Required) Tag name.
* `value` - (Required) Tag value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS Secrets Manager secret identifier and key, separated by a comma (`,`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_secretsmanager_tag` using the Organizations resource identifier and key, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_secretsmanager_tag.example
  id = "arn:aws:secretsmanager:us-east-1:123456789012:example-secret,ExampleKey"
}
```

Using `terraform import`, import `aws_secretsmanager_tag` using the AWS Secrets Manager secret identifier and key, separated by a comma (`,`). For example:

```console
% terraform import aws_secretsmanager_tag.example arn:aws:secretsmanager:us-east-1:123456789012:example-secret,ExampleKey
```
