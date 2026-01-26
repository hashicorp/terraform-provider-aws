---
subcategory: "AppStream 2.0"
layout: "aws"
page_title: "AWS: aws_appstream_directory_config"
description: |-
  Provides an AppStream Directory Config
---

# Resource: aws_appstream_directory_config

Provides an AppStream Directory Config.

## Example Usage

```terraform
resource "aws_appstream_directory_config" "example" {
  directory_name                          = "NAME OF DIRECTORY"
  organizational_unit_distinguished_names = ["DISTINGUISHED NAME"]

  service_account_credentials {
    account_name     = "NAME OF ACCOUNT"
    account_password = "PASSWORD OF ACCOUNT"
  }

  certificate_based_auth_properties {
    certificate_authority_arn = "ARN OF CERTIFICATE AUTHORITY"
    status                    = "STATUS OF CERTIFICATE BASED AUTHENTICATION"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `directory_name` - (Required) Fully qualified name of the directory.
* `organizational_unit_distinguished_names` - (Required) Distinguished names of the organizational units for computer accounts.
* `service_account_credentials` - (Required) Configuration block for the name of the directory and organizational unit (OU) to use to join the directory config to a Microsoft Active Directory domain. See [`service_account_credentials`](#service_account_credentials) below.
* `certificate_based_auth_properties` - (Optional) Configuration block for the certificate-based authentication properties used to authenticate SAML 2.0 Identity Provider (IdP) user identities to Active Directory domain-joined streaming instances. See [`certificate_based_auth_properties`](#certificate_based_auth_properties) below.

### `service_account_credentials`

* `account_name` - (Required) User name of the account. This account must have the following privileges: create computer objects, join computers to the domain, and change/reset the password on descendant computer objects for the organizational units specified.
* `account_password` - (Required) Password for the account.

### `certificate_based_auth_properties`

* `certificate_authority_arn` - (Optional) The ARN of the AWS Certificate Manager Private CA resource.
* `status` - (Optional) The status of the certificate-based authentication properties. Valid values - ["DISABLED", "ENABLED", "ENABLED_NO_DIRECTORY_LOGIN_FALLBACK"].

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier (ID) of the appstream directory config.
* `created_time` -  Date and time, in UTC and extended RFC 3339 format, when the directory config was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appstream_directory_config` using the id. For example:

```terraform
import {
  to = aws_appstream_directory_config.example
  id = "directoryNameExample"
}
```

Using `terraform import`, import `aws_appstream_directory_config` using the id. For example:

```console
% terraform import aws_appstream_directory_config.example directoryNameExample
```
