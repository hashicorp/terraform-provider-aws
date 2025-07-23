---
subcategory: "SES (Simple Email)"
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

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the domain identity.
* `domain` - Name of the domain
* `verification_token` - Code which when added to the domain as a TXT record will signal to SES that the owner of the domain has authorized SES to act on their behalf.
