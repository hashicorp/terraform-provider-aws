---
subcategory: "End User Messaging"
layout: "aws"
page_title: "AWS: aws_pinpoint_gcm_channel"
description: |-
  Provides an End User Messaging GCM Channel resource.
---

# Resource: aws_pinpoint_gcm_channel

Provides an End User Messaging GCM Channel resource.

~> **Note:** Credentials (Service Account JSON and API Key) will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```terraform
# Token method
resource "aws_pinpoint_gcm_channel" "gcm" {
  application_id                = aws_pinpoint_app.app.application_id
  default_authentication_method = "TOKEN"
  service_json                  = file("path_to_service_json")
}

# API Key (Legacy) method
resource "aws_pinpoint_gcm_channel" "gcm" {
  application_id                = aws_pinpoint_app.app.application_id
  default_authentication_method = "KEY"
  api_key                       = "api_key"
}

resource "aws_pinpoint_app" "app" {}
```

## Argument Reference

This resource supports the following arguments:

* `api_key` - (Optional) Platform credential API key from Google. Conflicts with `service_json`.
* `application_id` - (Required) Application ID.
* `default_authentication_method` - (Optional) Default authentication method used for GCM. Valid values: `KEY`, `TOKEN`. Defaults to `KEY`.
* `enabled` - (Optional) Whether the channel is enabled or disabled. Defaults to `true`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_json` - (Optional) Service Account JSON from Google to use with the GCM API. Conflicts with `api_key`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import End User Messaging GCM Channel using the `application-id`. For example:

```terraform
import {
  to = aws_pinpoint_gcm_channel.gcm
  id = "application-id"
}
```

Using `terraform import`, import End User Messaging GCM Channel using the `application-id`. For example:

```console
% terraform import aws_pinpoint_gcm_channel.gcm application-id
```
