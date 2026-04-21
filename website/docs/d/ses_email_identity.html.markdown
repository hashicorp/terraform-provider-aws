---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_email_identity"
description: |-
  Retrieve the active SES email identity
---

# Data Source: aws_ses_email_identity

Retrieve the active SES email identity

## Example Usage

```terraform
data "aws_ses_email_identity" "example" {
  email = "awesome@example.com"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` -  The ARN of the email identity.
* `email` - Email identity.
