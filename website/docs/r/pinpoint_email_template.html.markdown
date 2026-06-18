---
subcategory: "End User Messaging"
layout: "aws"
page_title: "AWS: aws_pinpoint_email_template"
description: |-
  Provides an End User Messaging Email Template resource.
---

# Resource: aws_pinpoint_email_template

~> **NOTE:** This resource is deprecated. AWS End User Messaging email features are being discontinued on October 30, 2026. Migrate to Amazon SES using [`aws_ses_template`](ses_template.html) or [`aws_sesv2_email_identity`](sesv2_email_identity.html) and related SESv2 resources. See the [AWS End User Messaging migration guide](https://docs.aws.amazon.com/pinpoint/latest/userguide/migrate.html) for details.

Provides an End User Messaging Email Template resource

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

This resource supports the following arguments:

* `email_template` - (Required, **Deprecated**) Content and settings for a message template that can be used in messages that are sent through the email channel. [See below](#email_template-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `template_name` - (Required, **Deprecated**) Name of the message template. A template name must start with an alphanumeric character and can contain a maximum of 128 characters. The characters can be alphanumeric characters, underscores (_), or hyphens (-). Template names are case sensitive.

### `email_template` Block

* `default_substitutions` - (Optional) JSON object that specifies the default values to use for message variables in the message template. This object is a set of key-value pairs. Each key defines a message variable in the template. The corresponding value defines the default value for that variable. When you create a message that's based on the template, you can override these defaults with message-specific and address-specific variables and values.
* `description` - (Optional) Custom description of the message template.
* `header` - (Optional) List of [MessageHeaders](https://docs.aws.amazon.com/pinpoint/latest/apireference/templates-template-name-email.html#templates-template-name-email-model-messageheader) for the email. You can have up to 15 Headers. [See below](#header-block).
* `html_part` - (Optional) Message body, in HTML format, to use in email messages that are based on the message template. We recommend using HTML format for email clients that render HTML content. You can include links, formatted text, and more in an HTML message.
* `recommender_id` - (Optional) Unique identifier for the recommender model to use for the message template. AWS End User Messaging uses this value to determine how to retrieve and process data from a recommender model when it sends messages that use the template, if the template contains message variables for recommendation data.
* `subject` - (Required) Subject line, or title, to use in email messages that are based on the message template.
* `text_part` - (Optional) Message body, in plain text format, to use in email messages that are based on the message template. We recommend using plain text format for email clients that don't render HTML content and clients that are connected to high-latency networks, such as mobile devices.

### `header` Block

* `name` - (Optional) Name of the message header. The header name can contain up to 126 characters.
* `value` - (Optional) Value of the message header. The header value can contain up to 870 characters, including the length of any rendered attributes. For example if you add the {CreationDate} attribute, it renders as YYYY-MM-DDTHH:MM:SS.SSSZ and is 24 characters in length.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the message template.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import End User Messaging Email Template using the `template_name`. For example:

```terraform
import {
  to = aws_pinpoint_email_template.reset
  id = "template_name"
}
```

Using `terraform import`, import End User Messaging Email Template using the `template_name`. For example:

```console
% terraform import aws_pinpoint_email_template.reset template_name
```
