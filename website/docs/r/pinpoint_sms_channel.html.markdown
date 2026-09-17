---
subcategory: "End User Messaging"
layout: "aws"
page_title: "AWS: aws_pinpoint_sms_channel"
description: |-
  Use the `aws_pinpoint_sms_channel` resource to manage End User Messaging SMS Channels.
---

# Resource: aws_pinpoint_sms_channel

Use the `aws_pinpoint_sms_channel` resource to manage End User Messaging SMS Channels.

~> **NOTE:** AWS has feature frozen the SMS channel of End User Messaging.

## Example Usage

```terraform
resource "aws_pinpoint_sms_channel" "sms" {
  application_id = aws_pinpoint_app.app.application_id
}

resource "aws_pinpoint_app" "app" {}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) ID of the application.
* `enabled` - (Optional) Whether the channel is enabled or disabled. By default, it is set to `true`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sender_id` - (Optional) Identifier of the sender for your messages.
* `short_code` - (Optional) Short Code registered with the phone provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `promotional_messages_per_second` - Maximum number of promotional messages that can be sent per second.
* `transactional_messages_per_second` - Maximum number of transactional messages per second that can be sent.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the End User Messaging SMS Channel using the `application_id`. For example:

```terraform
import {
  to = aws_pinpoint_sms_channel.sms
  id = "application-id"
}
```

Using `terraform import`, import the End User Messaging SMS Channel using the `application_id`. For example:

```console
% terraform import aws_pinpoint_sms_channel.sms application-id
```
