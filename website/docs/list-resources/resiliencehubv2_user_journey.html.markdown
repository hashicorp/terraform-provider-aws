---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_user_journey"
description: |-
  Lists Resilience Hub V2 User Journey resources for a system.
---

# List Resource: aws_resiliencehubv2_user_journey

Lists Resilience Hub V2 User Journey resources for a given system.

## Example Usage

### Basic

```terraform
list "aws_resiliencehubv2_user_journey" "example" {
  provider = aws

  config {
    system_arn = aws_resiliencehubv2_system.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to list resources in. Defaults to the provider's configured region.
* `system_arn` - (Required) ARN of the system whose user journeys are listed.
