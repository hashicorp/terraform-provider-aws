---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_application_integration_url"
description: |-
  Retrieves WAF Captcha Application Integration URL.
---
# Data Source: aws_wafv2_application_integration_url

Retrieves WAF Captcha Application Integration URL.

## Example Usage

```terraform
data "aws_wafv2_application_integration_url" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `url` - The Application Integration URL.
