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

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` -  The ARN of the email identity.
* `email` - Email identity.
