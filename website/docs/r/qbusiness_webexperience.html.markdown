---
subcategory: "Amazon Q Business"
layout: "aws"
page_title: "AWS: aws_qbusiness_webexperience"
description: |-
  Provides a Q Business Webexperience resource.
---

# Resource: aws_qbusiness_webexperience

Provides a Q Business Webexperience resource.

## Example Usage

```terraform
resource "aws_qbusiness_webexperience" "example" {
  application_id              = aws_qbusiness_app.this.id
  sample_prompts_control_mode = "ENABLED"
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) ID of the Amazon Q application.
* `sample_prompts_control_mode` - (Required) Sample prompts control mode for Amazon Q web experience. Valid values are `ENABLED` or `DISABLED`
* `subtitle` - (Optional) Subtitle for Amazon Q web experience.
* `title` - (Optional) Title for Amazon Q web experience.
* `welcome_message` - (Optional) Customized welcome message for end users of an Amazon Q web experience.
* `iam_service_role_arn` - (Optional) The Amazon Resource Name (ARN) of the service role attached to your web experience.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `webexperience_id` - ID of the Q Business Webexperience.
* `arn` - ARN of the Q Business Webexperience.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
