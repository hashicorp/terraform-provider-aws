---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_configuration_set"
description: |-
  Provides an SES configuration set
---

# Resource: aws_ses_configuration_set

Provides an SES configuration set resource

## Example Usage

```hcl
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-test"
}
```

### Require TLS Connections

```hcl
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-test"

  delivery_options {
    tls_policy = "Require"
  }
}
```

## Argument Reference

The following arguments are supported:

* `delivery_options` - (Optional) A configuration block that specifies whether messages that use the configuration set are required to use Transport Layer Security (TLS). Detailed below.
* `name` - (Required) The name of the configuration set

### delivery_options Argument Reference

The `delivery_options` configuration block supports the following argument:

* `tls_policy` - (Optional) Specifies whether messages that use the configuration set are required to use Transport Layer Security (TLS). If the value is `Require`, messages are only delivered if a TLS connection can be established. If the value is `Optional`, messages can be delivered in plain text if a TLS connection can't be established. Valid values: `Require` or `Optional`. Defaults to `Optional`.

## Import

SES Configuration Sets can be imported using their `name`, e.g.

```
$ terraform import aws_ses_configuration_set.test some-configuration-set-test
```
