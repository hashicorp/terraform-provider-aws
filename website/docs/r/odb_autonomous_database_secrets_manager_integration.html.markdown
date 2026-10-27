---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_autonomous_database_secrets_manager_integration"
page_title: "AWS: aws_odb_autonomous_database_secrets_manager_integration"
description: |-
  Manages the Oracle Database@AWS Autonomous Database Serverless integration with AWS Secrets Manager.
---

# Resource: aws_odb_autonomous_database_secrets_manager_integration

Manages the Oracle Database@AWS Autonomous Database Serverless integration with AWS Secrets Manager. The resource provisions an Oracle-managed service role that can assume a customer-managed role to read an administrator-password secret.

!> **Note:** This integration is shared by all databases in the AWS account and Region. Creating this resource manages the existing integration if it is already enabled. Destroying this resource disables the integration for the entire account and Region, which can disrupt databases outside this Terraform configuration that use AWS Secrets Manager credentials. Manage the integration in only one Terraform configuration and coordinate changes with all database owners that depend on it.

For accounts with a resource anchor, creating or destroying this resource preserves an existing OCI identity domain. If no identity domain exists, the service creates one when enabling or disabling the integration. This resource does not delete the identity domain.

Create the customer-managed IAM role separately. Its trust policy must allow the exported `role_arn` to assume it, and its permissions must grant access to the selected secret. See the [AWS Secrets Manager documentation](https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access.html) for IAM permission guidance.

## Example Usage

### Basic Usage

```terraform
resource "aws_odb_autonomous_database_secrets_manager_integration" "example" {}

output "adbs_secrets_manager_service_role_arn" {
  value = aws_odb_autonomous_database_secrets_manager_integration.example.role_arn
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Region where the integration is managed.
* `role_arn` - ARN of the Oracle-managed service role that assumes the customer-managed IAM role.
* `status` - Current lifecycle status of the service role.
* `status_reason` - Additional lifecycle-status information, if available.

## Timeouts

The `timeouts` configuration block supports the following arguments:

* `create` - (Default `15m`) Maximum time to wait for the service role to become available.
* `delete` - (Default `15m`) Maximum time to wait for the service role to be removed.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_odb_autonomous_database_secrets_manager_integration.example
  identity = {
    region = "us-east-1"
  }
}
```

### Identity Schema

#### Required

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the Region where the integration is managed. For example:

```terraform
import {
  to = aws_odb_autonomous_database_secrets_manager_integration.example
  id = "us-east-1"
}
```

Using `terraform import`, import the integration using its Region. For example:

```console
% terraform import aws_odb_autonomous_database_secrets_manager_integration.example us-east-1
```
