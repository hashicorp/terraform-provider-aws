---
subcategory: "End User Messaging"
layout: "aws"
page_title: "AWS: aws_pinpoint_apns_sandbox_channel"
description: |-
  Provides an End User Messaging APNs Sandbox Channel resource.
---

# Resource: aws_pinpoint_apns_sandbox_channel

Provides an End User Messaging APNs Sandbox Channel resource.

~> **Note:** All arguments, including certificates and tokens, will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```terraform
resource "aws_pinpoint_apns_sandbox_channel" "apns_sandbox" {
  application_id = aws_pinpoint_app.app.application_id

  certificate = file("./certificate.pem")
  private_key = file("./private_key.key")
}

resource "aws_pinpoint_app" "app" {}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) Application ID.
* `bundle_id` - (Required) ID assigned to your iOS app. To find this value, choose Certificates, IDs & Profiles, choose App IDs in the Identifiers section, and choose your app. Required if using Key credentials.
* `certificate` - (Required) Pem encoded TLS Certificate from Apple. Required if using Certificate credentials.
* `default_authentication_method` - (Optional) Default authentication method used for APNs Sandbox. __NOTE__: AWS End User Messaging uses this default for every APNs push notification that you send using the console. You can override the default when you send a message programmatically using the AWS End User Messaging API, the AWS CLI, or an AWS SDK. If your default authentication type fails, AWS End User Messaging doesn't attempt to use the other authentication type.
* `enabled` - (Optional) Whether the channel is enabled or disabled. Defaults to `true`.
* `private_key` - (Required) Certificate Private Key file (ie. `.key` file). Required if using Certificate credentials.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `team_id` - (Required) ID assigned to your Apple developer account team. This value is provided on the Membership page. Required if using Key credentials.
* `token_key` - (Required) `.p8` file that you download from your Apple developer account when you create an authentication key. Required if using Key credentials.
* `token_key_id` - (Required) ID assigned to your signing key. To find this value, choose Certificates, IDs & Profiles, and choose your key in the Keys section. Required if using Key credentials.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import End User Messaging APNs Sandbox Channel using the `application-id`. For example:

```terraform
import {
  to = aws_pinpoint_apns_sandbox_channel.apns_sandbox
  id = "application-id"
}
```

Using `terraform import`, import End User Messaging APNs Sandbox Channel using the `application-id`. For example:

```console
% terraform import aws_pinpoint_apns_sandbox_channel.apns_sandbox application-id
```
