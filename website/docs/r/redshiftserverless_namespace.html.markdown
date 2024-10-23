---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_namespace"
description: |-
  Provides a Redshift Serverless Namespace resource.
---

# Resource: aws_redshiftserverless_namespace

Creates a new Amazon Redshift Serverless Namespace.

## Example Usage

```terraform
resource "aws_redshiftserverless_namespace" "example" {
  namespace_name = "concurrency-scaling"
}
```

## Argument Reference

This resource supports the following arguments:

* `admin_password_secret_kms_key_id` - (Optional) ID of the KMS key used to encrypt the namespace's admin credentials secret.
* `admin_user_password` - (Optional) The password of the administrator for the first database created in the namespace.
  Conflicts with `manage_admin_password`.
* `admin_username` - (Optional) The username of the administrator for the first database created in the namespace.
* `db_name` - (Optional) The name of the first database created in the namespace.
* `default_iam_role_arn` - (Optional) The Amazon Resource Name (ARN) of the IAM role to set as a default in the namespace. When specifying `default_iam_role_arn`, it also must be part of `iam_roles`.
* `iam_roles` - (Optional) A list of IAM roles to associate with the namespace.
* `kms_key_id` - (Optional) The ARN of the Amazon Web Services Key Management Service key used to encrypt your data.
* `log_exports` - (Optional) The types of logs the namespace can export. Available export types are `userlog`, `connectionlog`, and `useractivitylog`.
* `namespace_name` - (Required) The name of the namespace.
* `manage_admin_password` - (Optional) Whether to use AWS SecretManager to manage namespace's admin credentials.
  Conflicts with `admin_user_password`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Redshift Serverless Namespace.
* `id` - The Redshift Namespace Name.
* `namespace_id` - The Redshift Namespace ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Serverless Namespaces using the `namespace_name`. For example:

```terraform
import {
  to = aws_redshiftserverless_namespace.example
  id = "example"
}
```

Using `terraform import`, import Redshift Serverless Namespaces using the `namespace_name`. For example:

```console
% terraform import aws_redshiftserverless_namespace.example example
```
