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
* `saml_options` - (Required) Configuration block for SAML options.
* `type` - (Required) Type of configuration. Must be `saml`.

The following arguments are optional:

* `description` - (Optional) Description of the security configuration.

### saml_options

* `group_attribute` - (Optional) Group attribute for this SAML integration.
* `metadata` - (Required) The XML IdP metadata file generated from your identity provider.
* `session_timeout` - (Optional) Session timeout, in minutes. Minimum is 5 minutes and maximum is 720 minutes (12 hours). Default is 60 minutes.
* `user_attribute` - (Optional) User attribute for this SAML integration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `config_version` - Version of the configuration.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearchServerless Access Policy using the `name` argument prefixed with the string `saml/account_id/`. For example:

```terraform
import {
  to = aws_opensearchserverless_security_config.example
  id = "saml/123456789012/example"
}
```

Using `terraform import`, import OpenSearchServerless Access Policy using the `name` argument prefixed with the string `saml/account_id/`. For example:

```console
% terraform import aws_opensearchserverless_security_config.example saml/123456789012/example
```
