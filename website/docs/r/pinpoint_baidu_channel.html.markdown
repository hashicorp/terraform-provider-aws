---
subcategory: "Pinpoint"
layout: "aws"
page_title: "AWS: aws_pinpoint_baidu_channel"
description: |-
  Provides a Pinpoint Baidu Channel resource.
---

# Resource: aws_pinpoint_baidu_channel

Provides a Pinpoint Baidu Channel resource.

~> **Note:** All arguments including the Api Key and Secret Key will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```terraform
resource "aws_pinpoint_app" "app" {}

resource "aws_pinpoint_baidu_channel" "channel" {
  application_id = aws_pinpoint_app.app.application_id
  api_key        = ""
  secret_key     = ""
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) The application ID.
* `enabled` - (Optional) Specifies whether to enable the channel. Defaults to `true`.
* `api_key` - (Required) Platform credential API key from Baidu.
* `secret_key` - (Required) Platform credential Secret key from Baidu.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Pinpoint Baidu Channel using the `application-id`. For example:

```terraform
import {
  to = aws_pinpoint_baidu_channel.channel
  id = "application-id"
}
```

Using `terraform import`, import Pinpoint Baidu Channel using the `application-id`. For example:

```console
% terraform import aws_pinpoint_baidu_channel.channel application-id
```
