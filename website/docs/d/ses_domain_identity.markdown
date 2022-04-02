---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_domain_identity"
description: |-
  Retrieve the SES domain identity
---

# Data Source: aws_ses_domain_identity

Retrieve the SES domain identity

## Example Usage

```terraform
data "aws_ses_domain_identity" "example" {
  domain = "example.com"
}
```

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN of the domain identity.
* `domain` - The name of the domain
* `verification_token` - A code which when added to the domain as a TXT record will signal to SES that the owner of the domain has authorized SES to act on their behalf.
