---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_event_bus"
description: |-
  Lists EventBridge V2 event buses.
---

# List Resource: aws_eventbridgev2_event_bus

Lists EventBridge V2 event buses. Only event buses owned by the caller's account are listed; event buses shared with the account through AWS RAM are excluded.

## Example Usage

```terraform
list "aws_eventbridgev2_event_bus" "example" {
  provider = aws

  config {
    name_prefix = "example"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `name_prefix` - (Optional) Name prefix used to filter event buses.
* `region` - (Optional) Region to query. Defaults to provider region.
