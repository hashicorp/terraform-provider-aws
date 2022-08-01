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

The following arguments are supported:

* `admin_user_password` - (Optional) The password of the administrator for the first database created in the namespace.
* `admin_username` - (Optional) The username of the administrator for the first database created in the namespace.
* `db_name` - (Optional) The name of the first database created in the namespace.
* `default_iam_role_arn` - (Optional) The Amazon Resource Name (ARN) of the IAM role to set as a default in the namespace.
* `iam_roles` - (Optional) A list of IAM roles to associate with the namespace.
* `kms_key_id` - (Optional) The ARN of the Amazon Web Services Key Management Service key used to encrypt your data.
* `log_exports` - (Optional) The types of logs the namespace can export. Available export types are `userlog`, `connectionlog`, and `useractivitylog`.
* `namespace_name` - (Required) The name of the namespace.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Redshift Serverless Namespace.
* `id` - The Redshift Namespace Name.
* `namespace_id` - The Redshift Namespace ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Redshift Servreless Namespaces can be imported using the `namespace_name`, e.g.,

```
$ terraform import aws_redshiftserverless_namespace.example example
```
