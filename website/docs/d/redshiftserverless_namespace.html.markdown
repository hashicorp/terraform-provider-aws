---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_namespace"
description: |-
  Terraform data source for managing an AWS Redshift Serverless Namespace.
---

# Data Source: aws_redshiftserverless_namespace

Terraform data source for managing an AWS Redshift Serverless Namespace.

## Example Usage

```terraform
data "aws_redshiftserverless_namespace" "example" {
  namespace_name = "example-namespace"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `namespace_name` - (Required) The name of the namespace.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `admin_username` - The username of the administrator for the first database created in the namespace.
* `arn` - ARN of the Redshift Serverless Namespace.
* `db_name` - The name of the first database created in the namespace.
* `default_iam_role_arn` - ARN of the IAM role to set as a default in the namespace. When specifying `default_iam_role_arn`, it also must be part of `iam_roles`.
* `iam_roles` - A list of IAM roles to associate with the namespace.
* `kms_key_id` - ARN of the Amazon Web Services KMS key used to encrypt your data.
* `log_exports` - The types of logs the namespace can export. Available export types are `userlog`, `connectionlog`, and `useractivitylog`.
* `namespace_id` - The Redshift Namespace ID.
