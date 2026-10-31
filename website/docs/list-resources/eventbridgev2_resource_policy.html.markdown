---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_resource_policy"
description: |-
  Lists EventBridge V2 resource policies.
---

# List Resource: aws_eventbridgev2_resource_policy

Lists EventBridge V2 resource policies. Only the customer-managed (`default`) policies of event buses owned by the caller's account are listed.

## Example Usage

```terraform
list "aws_eventbridgev2_resource_policy" "example" {
  provider = aws

  config {
    name_prefix = "example"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `name_prefix` - (Optional) Name prefix used to filter the event buses whose policies are listed.
* `region` - (Optional) Region to query. Defaults to provider region.
