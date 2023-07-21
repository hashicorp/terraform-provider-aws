---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_spot_datafeed_subscription"
description: |-
  Provides a Spot Datafeed Subscription resource.
---

# Resource: aws_spot_datafeed_subscription

-> **Note:** There is only a single subscription allowed per account.

To help you understand the charges for your Spot instances, Amazon EC2 provides a data feed that describes your Spot instance usage and pricing.
This data feed is sent to an Amazon S3 bucket that you specify when you subscribe to the data feed.

## Example Usage

```terraform
resource "aws_s3_bucket" "default" {
  bucket = "tf-spot-datafeed"
}

resource "aws_spot_datafeed_subscription" "default" {
  bucket = aws_s3_bucket.default.id
  prefix = "my_subdirectory"
}
```

## Argument Reference

* `bucket` - (Required) The Amazon S3 bucket in which to store the Spot instance data feed.
* `prefix` - (Optional) Path of folder inside bucket to place spot pricing data.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Spot Datafeed Subscription using the word `spot-datafeed-subscription`. For example:

```terraform
import {
  to = aws_spot_datafeed_subscription.mysubscription
  id = "spot-datafeed-subscription"
}
```

Using `terraform import`, import a Spot Datafeed Subscription using the word `spot-datafeed-subscription`. For example:

```console
% terraform import aws_spot_datafeed_subscription.mysubscription spot-datafeed-subscription
```
