---
subcategory: "Security Incident Response"
layout: "aws"
page_title: "AWS: aws_securityir_membership"
description: |-
  Lists Security Incident Response Membership resources.
---

# List Resource: aws_securityir_membership

Lists Security Incident Response Membership resources. Only active memberships in the queried Region are returned.

## Example Usage

```terraform
list "aws_securityir_membership" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
