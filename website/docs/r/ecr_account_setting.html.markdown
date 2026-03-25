---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_account_setting"
description: |-
  Provides a resource to manage AWS ECR account settings
---

# Resource: aws_ecr_account_setting

Provides a resource to manage AWS ECR account settings

## Example Usage

### Configuring Basic Scanning

```terraform
resource "aws_ecr_account_setting" "basic_scan_type_version" {
  name  = "BASIC_SCAN_TYPE_VERSION"
  value = "AWS_NATIVE"
}
```

### Configuring Blob Mounting (Cross-Repository Layer Sharing)

```terraform
resource "aws_ecr_account_setting" "blob_mounting" {
  name  = "BLOB_MOUNTING"
  value = "ENABLED"
}
```

### Configuring Registry Policy Scope

```terraform
resource "aws_ecr_account_setting" "registry_policy_scope" {
  name  = "REGISTRY_POLICY_SCOPE"
  value = "V2"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the account setting. One of: `BASIC_SCAN_TYPE_VERSION`, `BLOB_MOUNTING`, `REGISTRY_POLICY_SCOPE`.
* `value` - (Required) Setting value that is specified. Valid values are:
    * If `name` is specified as `BASIC_SCAN_TYPE_VERSION`, one of: `AWS_NATIVE`, `CLAIR`.
    * If `name` is specified as `BLOB_MOUNTING`, one of: `ENABLED`, `DISABLED`.
    * If `name` is specified as `REGISTRY_POLICY_SCOPE`, one of: `V1`, `V2`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name of the account setting.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECR Scan Type using the `name`. For example:

```terraform
import {
  to = aws_ecr_account_setting.foo
  id = "BASIC_SCAN_TYPE_VERSION"
}
```

Using `terraform import`, import EMR Security Configurations using the account setting name. For example:

```console
% terraform import aws_ecr_account_setting.foo BASIC_SCAN_TYPE_VERSION
```
