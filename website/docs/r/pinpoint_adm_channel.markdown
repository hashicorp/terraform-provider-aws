---
subcategory: "Pinpoint"
layout: "aws"
page_title: "AWS: aws_pinpoint_adm_channel"
description: |-
  Provides a Pinpoint ADM Channel resource.
---

# Resource: aws_pinpoint_adm_channel

Provides a Pinpoint ADM (Amazon Device Messaging) Channel resource.

~> **Note:** All arguments including the Client ID and Client Secret will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```terraform
resource "aws_pinpoint_app" "app" {}

resource "aws_pinpoint_adm_channel" "channel" {
  application_id = aws_pinpoint_app.app.application_id
  client_id      = ""
  client_secret  = ""
  enabled        = true
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) The application ID.
* `client_id` - (Required) Client ID (part of OAuth Credentials) obtained via Amazon Developer Account.
* `client_secret` - (Required) Client Secret (part of OAuth Credentials) obtained via Amazon Developer Account.
* `enabled` - (Optional) Specifies whether to enable the channel. Defaults to `true`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Pinpoint ADM Channel using the `application-id`. For example:

```terraform
import {
  to = aws_pinpoint_adm_channel.channel
  id = "application-id"
}
```

Using `terraform import`, import Pinpoint ADM Channel using the `application-id`. For example:

```console
% terraform import aws_pinpoint_adm_channel.channel application-id
```
