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

The following arguments are required:

* `id` - (Required) The unique identifier of the security configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `config_version` - The version of the security configuration.
* `created_date` - The date the configuration was created.
* `description` - The description of the security configuration.
* `last_modified_date` - The date the configuration was last modified.
* `saml_options` - SAML options for the security configuration.
* `type` - The type of security configuration.

### saml_options

SAML options for the security configuration.

* `group_attribute` - Group attribute for this SAML integration.
* `metadata` - The XML IdP metadata file generated from your identity provider.
* `session_timeout` - Session timeout, in minutes. Minimum is 5 minutes and maximum is 720 minutes (12 hours). Default is 60 minutes.
* `user_attribute` - User attribute for this SAML integration.
