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

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `url` - The Application Integration URL.
