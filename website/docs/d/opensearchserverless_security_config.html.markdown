---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_security_config"
description: |-
  Terraform data source for managing an AWS OpenSearch Serverless Security Config.
---

# Data Source: aws_opensearchserverless_security_config

Terraform data source for managing an AWS OpenSearch Serverless Security Config.

## Example Usage

### Basic Usage

```terraform
data "aws_opensearchserverless_security_config" "example" {
  id = "saml/12345678912/example"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Required) Unique identifier of the security configuration.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `config_version` - Version of the security configuration.
* `created_date` - Date the configuration was created.
* `description` - Description of the security configuration.
* `iam_federation_options` - IAM Federation options for the security configuration.
* `iam_identity_center_options` - IAM Identity Center options for the security configuration.
* `last_modified_date` - Date the configuration was last modified.
* `saml_options` - SAML options for the security configuration.
* `type` - Type of security configuration.

### `iam_federation_options` Block

IAM Federation options for the security configuration.

* `group_attribute` - Group attribute for this IAM federation integration.
* `user_attribute` - User attribute for this IAM federation integration.

### `iam_identity_center_options` Block

IAM Identity Center options for the security configuration.

* `group_attribute` - Group attribute for this IAM Identity Center integration.
* `instance_arn` - Amazon Resource Name (ARN) of the IAM Identity Center instance used to integrate with OpenSearch Serverless.
* `user_attribute` - User attribute for this IAM Identity Center integration.

### `saml_options` Block

SAML options for the security configuration.

* `group_attribute` - Group attribute for this SAML integration.
* `metadata` - XML IdP metadata file generated from your identity provider.
* `session_timeout` - Session timeout, in minutes. Minimum is 5 minutes and maximum is 720 minutes (12 hours). Default is 60 minutes.
* `user_attribute` - User attribute for this SAML integration.
