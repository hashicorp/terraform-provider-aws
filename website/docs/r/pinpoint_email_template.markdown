---
subcategory: "Pinpoint"
layout: "aws"
page_title: "AWS: aws_pinpoint_email_template"
description: |-
  Provides a Pinpoint Email Template resource.
---

# Resource: aws_pinpoint_email_template

Provides a Pinpoint Email Template resource

## Example Usage

```terraform
resource "aws_pinpoint_email_template" "password_reset" {
  name        = "PasswordReset"
  description = "Used for password reset emails"
  html        = file("${path.module}/templates/password_reset_email.tpl")
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the message template. A template name must start with an alphanumeric character and can contain a maximum of 128 characters. The characters can be alphanumeric characters, underscores (_), or hyphens (-). Template names are case sensitive.
* `html` - (Optional) The message body, in HTML format, to use in email messages that are based on the message template. We recommend using HTML format for email clients that render HTML content. You can include links, formatted text, and more in an HTML message.
* `text` - (Optional) The message body, in plain text format, to use in email messages that are based on the message template. We recommend using plain text format for email clients that don't render HTML content and clients that are connected to high-latency networks, such as mobile devices.
* `subject` - (Optional) The subject line, or title, to use in email messages that are based on the message template.
* `default_substitutions` - (Optional) A JSON object that specifies the default values to use for message variables in the message template. This object is a set of key-value pairs. Each key defines a message variable in the template. The corresponding value defines the default value for that variable. When you create a message that's based on the template, you can override these defaults with message-specific and address-specific variables and values.
* `recommender_id` - (Optional) The unique identifier for the recommender model to use for the message template. Amazon Pinpoint uses this value to determine how to retrieve and process data from a recommender model when it sends messages that use the template, if the template contains message variables for recommendation data.
* `description` - (Optional) A custom description of the message template.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the message template.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Pinpoint Email Template using the `name`. For example:

```terraform
import {
  to = aws_pinpoint_email_template.reset
  id = "name"
}
```

Using `terraform import`, import Pinpoint Email Template using the `name`. For example:

```console
% terraform import aws_pinpoint_email_template.reset name
```
