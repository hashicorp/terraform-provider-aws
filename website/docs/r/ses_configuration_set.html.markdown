---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_configuration_set"
description: |-
  Provides an SES configuration set
---

# Resource: aws_ses_configuration_set

Provides an SES configuration set resource.

## Example Usage

### Basic Example

```terraform
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-test"
}
```

### Require TLS Connections

```terraform
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-test"

  delivery_options {
    tls_policy = "Require"
  }
}
```

### Tracking Options

```terraform
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-test"

  tracking_options {
    custom_redirect_domain = "sub.example.com"
  }
}
```

## Argument Reference

The following argument is required:

* `name` - (Required) Name of the configuration set.

The following argument is optional:

* `delivery_options` - (Optional) Whether messages that use the configuration set are required to use TLS. See below.
* `reputation_metrics_enabled` - (Optional) Whether or not Amazon SES publishes reputation metrics for the configuration set, such as bounce and complaint rates, to Amazon CloudWatch. The default value is `false`.
* `sending_enabled` - (Optional) Whether email sending is enabled or disabled for the configuration set. The default value is `true`.
* `tracking_options` - (Optional) Domain that is used to redirect email recipients to an Amazon SES-operated domain. See below. **NOTE:** This functionality is best effort.

### delivery_options

* `tls_policy` - (Optional) Whether messages that use the configuration set are required to use Transport Layer Security (TLS). If the value is `Require`, messages are only delivered if a TLS connection can be established. If the value is `Optional`, messages can be delivered in plain text if a TLS connection can't be established. Valid values: `Require` or `Optional`. Defaults to `Optional`.

### tracking_options

* `custom_redirect_domain` - (Optional) Custom subdomain that is used to redirect email recipients to the Amazon SES event tracking domain.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - SES configuration set ARN.
* `id` - SES configuration set name.
* `last_fresh_start` - Date and time at which the reputation metrics for the configuration set were last reset. Resetting these metrics is known as a fresh start.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SES Configuration Sets using their `name`. For example:

```terraform
import {
  to = aws_ses_configuration_set.test
  id = "some-configuration-set-test"
}
```

Using `terraform import`, import SES Configuration Sets using their `name`. For example:

```console
% terraform import aws_ses_configuration_set.test some-configuration-set-test
```
