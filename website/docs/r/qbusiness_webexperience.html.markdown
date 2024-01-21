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
  application_id               = aws_qbusiness_app.this.id
  sample_propmpts_control_mode = "ENABLED"
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) ID of the Amazon Q application.
* `authentication_configuration` - (Optional) Authentication configuration of the Amazon Q web experience.
* `sample_propmpts_control_mode` - (Optional) Sample prompts control mode for Amazon Q web experience.
* `subtitle` - (Optional) Subtitle for Amazon Q web experience.
* `title` - (Optional) Title for Amazon Q web experience.
* `welcome_message` - (Optional) Customized welcome message for end users of an Amazon Q web experience.

`authentication_configuration` supports the following:

* `saml_configuration` - (Required) Status information about whether file upload functionality is activated or deactivated for your end user. Valid values are `ENABLED` and `DISABLED`.

`saml_configuration` supports the following:

* `metadata_xml` - (Required) SAML metadata document provided by your identity provider (IdP).
* `iam_role_arn` - (Required) ARN of an IAM role assumed by users when they authenticate into their Amazon Q web experience, containing the relevant Amazon Q permissions for conversing with Amazon Q.
* `user_id_attribute` - (Required) User attribute name in your IdP that maps to the user email.
* `user_group_attribute` - (Optional) The SAML metadata document provided by your identity provider (IdP).


## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `webexperience_id` - ID of the Q Business Webexperience.
* `arn` - ARN of the Q Business Webexperience.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
