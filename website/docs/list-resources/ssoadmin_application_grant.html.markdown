---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application_grant"
description: |-
  Lists SSO Admin Application Grant resources.
---

# List Resource: aws_ssoadmin_application_grant

Lists SSO Admin Application Grant resources.

## Example Usage

```terraform
list "aws_ssoadmin_application_grant" "example" {
  provider = aws

  config {
    application_arn = aws_ssoadmin_application.example.application_arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `application_arn` - (Required) ARN of the application whose grants to list.
* `region` - (Optional) Region to query. Defaults to provider region.
