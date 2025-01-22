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
resource "aws_pinpoint_email_template" "test" {
  template_name = "testing"
  email_template {
    subject   = "testing"
    text_part = "we are testing template text part"
    header {
      name  = "testingname"
      value = "testingvalue"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `template_name` - (Required) name of the message template. A template name must start with an alphanumeric character and can contain a maximum of 128 characters. The characters can be alphanumeric characters, underscores (_), or hyphens (-). Template names are case sensitive.
* `email_template` - (Required) Specifies the content and settings for a message template that can be used in messages that are sent through the email channel. See [Email Template](#email-template)

### Email Template

* `default_substitutions` - (Optional) JSON object that specifies the default values to use for message variables in the message template. This object is a set of key-value pairs. Each key defines a message variable in the template. The corresponding value defines the default value for that variable. When you create a message that's based on the template, you can override these defaults with message-specific and address-specific variables and values.
* `headers` - (Required) List of [MessageHeaders](https://docs.aws.amazon.com/pinpoint/latest/apireference/templates-template-name-email.html#templates-template-name-email-model-messageheader) for the email. You can have up to 15 Headers. See [Headers](#headers)
* `html_part` - (Optional) The message body, in HTML format, to use in email messages that are based on the message template. We recommend using HTML format for email clients that render HTML content. You can include links, formatted text, and more in an HTML message.
* `recommender_id` - (Optional) The unique identifier for the recommender model to use for the message template. Amazon Pinpoint uses this value to determine how to retrieve and process data from a recommender model when it sends messages that use the template, if the template contains message variables for recommendation data.
* `subject` - (Required) Subject line, or title, to use in email messages that are based on the message template.
* `tags` - *Deprecated* As of 22-05-2023 tags has been deprecated for update operations. After this date any value in tags is not processed and an error code is not returned.
* `template_description` - (Optional) Custom description of the message template.
* `text_part` - (Optional) Message body, in plain text format, to use in email messages that are based on the message template. We recommend using plain text format for email clients that don't render HTML content and clients that are connected to high-latency networks, such as mobile devices.

### Headers

* `name` - Name of the message header. The header name can contain up to 126 characters.
* `value` - Value of the message header. The header value can contain up to 870 characters, including the length of any rendered attributes. For example if you add the {CreationDate} attribute, it renders as YYYY-MM-DDTHH:MM:SS.SSSZ and is 24 characters in length.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the message template.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Pinpoint Email Template using the `template_name`. For example:

```terraform
import {
  to = aws_pinpoint_email_template.reset
  id = "template_name"
}
```

Using `terraform import`, import Pinpoint Email Template using the `template_name`. For example:

```console
% terraform import aws_pinpoint_email_template.reset template_name
```
