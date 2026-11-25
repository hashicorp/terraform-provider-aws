---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_service_function"
description: |-
  Lists Resilience Hub V2 Service Function resources for a service.
---

# List Resource: aws_resiliencehubv2_service_function

Lists Resilience Hub V2 Service Function resources for a given service.

## Example Usage

### Basic

```terraform
list "aws_resiliencehubv2_service_function" "example" {
  provider = aws

  config {
    service_arn = aws_resiliencehubv2_service.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to list resources in. Defaults to the provider's configured region.
* `service_arn` - (Required) ARN of the service whose service functions are listed.
