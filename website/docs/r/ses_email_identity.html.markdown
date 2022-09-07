---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_email_identity"
description: |-
  Provides an SES email identity resource
---

# Resource: aws_ses_email_identity

Provides an SES email identity resource

## Argument Reference

The following arguments are supported:

* `email` - (Required) The email address to assign to SES
* `default_configuration_set` - (Optional) The configuration set name to use as default for this identity. see [here](https://docs.aws.amazon.com/ses/latest/dg/managing-identities.html#managing-configuration-sets-default-adding) for more info
* `tags` - (Optional) Map of tags to assign to the email identity. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the email identity.

## Example Usage

```terraform
resource "aws_ses_email_identity" "example" {
  email = "email@example.com"
  default_configuration_set = "example-cfg-set"
  tags = {
    Type = "Identity"
  }
}
```

## Import

SES email identities can be imported using the email address.

```
$ terraform import aws_ses_email_identity.example email@example.com
```
