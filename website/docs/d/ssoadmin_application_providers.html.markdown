---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application_providers"
description: |-
  Terraform data source for managing AWS SSO Admin Application Providers.
---

# Data Source: aws_ssoadmin_application_providers

Terraform data source for managing AWS SSO Admin Application Providers.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_application_providers" "example" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS region.
* `application_providers` - A list of application providers available in the current region. See [`application_providers`](#application_providers-attribute-reference) below.

### `application_providers` Attribute Reference

* `application_provider_arn` - ARN of the application provider.
* `display_data` - An object describing how IAM Identity Center represents the application provider in the portal. See [`display_data`](#display_data-attribute-reference) below.
* `federation_protocol` - Protocol that the application provider uses to perform federation. Valid values are `SAML` and `OAUTH`.

### `display_data` Attribute Reference

* `description` - Description of the application provider.
* `display_name` - Name of the application provider.
* `icon_url` - URL that points to an icon that represents the application provider.
