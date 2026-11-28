---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_event_destination"
description: |-
  Lists AWS End User Messaging SMS Event Destination resources.
---

# List Resource: aws_pinpointsmsvoicev2_event_destination

Lists AWS End User Messaging SMS Event Destination resources.

## Example Usage

```terraform
list "aws_pinpointsmsvoicev2_event_destination" "example" {
  provider = aws

  config {
    configuration_set_names = ["my-configuration-set"]
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `configuration_set_names` - (Required) Names of configuration sets to list event destinations for.
* `region` - (Optional) Region to query. Defaults to provider region.
