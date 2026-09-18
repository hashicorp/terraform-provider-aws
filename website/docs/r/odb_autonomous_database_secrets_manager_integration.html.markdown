---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_autonomous_database_secrets_manager_integration"
page_title: "AWS: aws_odb_autonomous_database_secrets_manager_integration"
description: |-
  Terraform resource for enabling the Oracle Database@AWS Autonomous Database Serverless integration with AWS Secrets Manager.
---

# Resource: aws_odb_autonomous_database_secrets_manager_integration

Enables Oracle Database@AWS Autonomous Database Serverless to use AWS Secrets Manager credentials. The resource provisions an Oracle-managed service role that can assume a customer-managed role to read an administrator-password secret.

Create the customer-managed IAM role separately. Its trust policy must allow the exported `role_arn` to assume it, and its permissions must grant access to the selected secret. See the [AWS Secrets Manager documentation](https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access.html) for IAM permission guidance.

## Example Usage

```terraform
resource "aws_odb_autonomous_database_secrets_manager_integration" "example" {}

output "adbs_secrets_manager_service_role_arn" {
  value = aws_odb_autonomous_database_secrets_manager_integration.example.role_arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier for the integration.
* `role_arn` - ARN of the Oracle-managed service role that assumes the customer-managed IAM role.
* `status` - Current lifecycle status of the service role.
* `status_reason` - Additional lifecycle-status information, if available.

## Timeouts

The `timeouts` configuration block supports the following arguments:

* `create` - (Default `15m`) Maximum time to wait for the service role to become available.
* `delete` - (Default `15m`) Maximum time to wait for the service role to be removed.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `secrets-manager` identifier. For example:

```terraform
import {
  to = aws_odb_autonomous_database_secrets_manager_integration.example
  id = "secrets-manager"
}
```

Using `terraform import`, import the integration using the `secrets-manager` identifier. For example:

```console
% terraform import aws_odb_autonomous_database_secrets_manager_integration.example secrets-manager
```
