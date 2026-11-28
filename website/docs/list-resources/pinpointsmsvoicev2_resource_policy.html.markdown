---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_resource_policy"
description: |-
  Lists End User Messaging SMS Resource Policy resources.
---

# List Resource: aws_pinpointsmsvoicev2_resource_policy

Lists the resource-based policy attached to an End User Messaging SMS resource.

As a resource policy is retrieved per parent resource (there is no bulk list API), listing is scoped to a single `resource_arn` and returns the attached policy, if any.

## Example Usage

```terraform
list "aws_pinpointsmsvoicev2_resource_policy" "example" {
  provider = aws

  config {
    resource_arn = aws_pinpointsmsvoicev2_phone_number.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `resource_arn` - (Required) ARN of the End User Messaging SMS resource whose attached policy to list.
