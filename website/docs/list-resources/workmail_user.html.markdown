---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_user"
description: |-
  Lists WorkMail User resources.
---

# List Resource: aws_workmail_user

Lists WorkMail User resources.

## Example Usage

```terraform
list "aws_workmail_user" "example" {
  provider = aws

  config {
    organization_id = aws_workmail_organization.example.organization_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `organization_id` - (Required) ID of the WorkMail organization to list users from.
* `region` - (Optional) Region to query. Defaults to provider region.
