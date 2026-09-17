---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_domain"
description: |-
  Lists WorkMail Domain resources.
---

# List Resource: aws_workmail_domain

Lists WorkMail Domain resources.

## Example Usage

```terraform
list "aws_workmail_domain" "example" {
  provider = aws

  config {
    organization_id = aws_workmail_organization.example.organization_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `organization_id` - (Required) ID of the WorkMail organization to list domains from.
* `region` - (Optional) Region to query. Defaults to provider region.
