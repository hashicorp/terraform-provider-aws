---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_group"
description: |-
  Lists WorkMail Group resources.
---

# List Resource: aws_workmail_group

Lists WorkMail Group resources.

## Example Usage

```terraform
list "aws_workmail_group" "example" {
  provider = aws

  config {
    organization_id = aws_workmail_organization.example.organization_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `organization_id` - (Required) ID of the WorkMail organization to list groups from.
* `region` - (Optional) Region to query. Defaults to provider region.
