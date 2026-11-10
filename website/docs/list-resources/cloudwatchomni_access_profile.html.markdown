---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_access_profile"
description: |-
  Lists CloudWatch Omni Access Profile resources.
---

# List Resource: aws_cloudwatchomni_access_profile

Lists CloudWatch Omni Access Profile resources in a space. Service-managed access profiles are not listed.

## Example Usage

```terraform
list "aws_cloudwatchomni_access_profile" "example" {
  provider = aws

  config {
    space_id = aws_cloudwatchomni_space.example.space_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `space_id` - (Required) ID of the space whose access profiles are listed.
