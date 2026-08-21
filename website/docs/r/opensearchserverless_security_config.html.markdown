---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_security_config"
description: |-
  Terraform resource for managing an AWS OpenSearch Serverless Security Config.
---

# Resource: aws_opensearchserverless_security_config

Terraform resource for managing an AWS OpenSearch Serverless Security Config.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearchserverless_security_config" "example" {
  name = "example"
  type = "saml"
  saml_options {
    metadata = file("${path.module}/idp-metadata.xml")
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Name of the policy.
* `type` - (Required) Type of configuration. Valid values are `saml`, `iamidentitycenter` and `iamfederation`.

The following arguments are optional:

* `description` - (Optional) Description of the security configuration.
* `iam_federation_options` - (Optional) Configuration block for IAM Federation options. Required if `type` is set to `iamfederation`. See [`iam_federation_options` Block](#iam_federation_options-block) below for details.
* `iam_identity_center_options` - (Optional) Configuration block for IAM Identity Center options. Required if `type` is set to `iamidentitycenter`. See [`iam_identity_center_options` Block](#iam_identity_center_options-block) below for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `saml_options` - (Optional) Configuration block for SAML options. Required if `type` is set to `saml`. See [`saml_options` Block](#saml_options-block) below for details.

### `iam_federation_options` Block

* `group_attribute` - (Optional) Group attribute for this IAM federation integration. At least one of `group_attribute` or `user_attribute` must be specified.
* `user_attribute` - (Optional) User attribute for this IAM federation integration. At least one of `group_attribute` or `user_attribute` must be specified.

### `iam_identity_center_options` Block

* `group_attribute` - (Optional) Group attribute for this IAM Identity Center integration. Valid values are `GroupId` and `GroupName`. Defaults to `GroupId`.
* `instance_arn` - (Required, Forces new resource) Amazon Resource Name (ARN) of the IAM Identity Center instance used to integrate with OpenSearch Serverless.
* `user_attribute` - (Optional) User attribute for this IAM Identity Center integration. Valid values are `UserId`, `UserName` and `Email`. Defaults to `UserId`.

### `saml_options` Block

* `group_attribute` - (Optional) Group attribute for this SAML integration.
* `metadata` - (Required) XML IdP metadata file generated from your identity provider.
* `session_timeout` - (Optional) Session timeout, in minutes. Minimum is 5 minutes and maximum is 720 minutes (12 hours). Default is 60 minutes.
* `user_attribute` - (Optional) User attribute for this SAML integration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `config_version` - Version of the configuration.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_opensearchserverless_security_config.example
  identity = {
    name = "example"
    type = "saml"
  }
}

resource "aws_opensearchserverless_security_config" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the policy.
* `type` (String) Type of configuration.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearchServerless Security Config using the `name` argument prefixed with the string `saml/account_id/`. For example:

```terraform
import {
  to = aws_opensearchserverless_security_config.example
  id = "saml/123456789012/example"
}
```

Using `terraform import`, import OpenSearchServerless Security Config using the `name` argument prefixed with the string `saml/account_id/`. For example:

```console
% terraform import aws_opensearchserverless_security_config.example saml/123456789012/example
```
