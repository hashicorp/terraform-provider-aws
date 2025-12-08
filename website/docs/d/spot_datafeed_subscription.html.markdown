---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_spot_datafeed_subscription"
description: |-
  Terraform data source for accessing an AWS EC2 (Elastic Compute Cloud) spot data feed subscription.
---

# Data Source: aws_spot_datafeed_subscription

~> There is only a single spot data feed subscription per account.

Terraform data source for accessing an AWS EC2 (Elastic Compute Cloud) spot data feed subscription.

## Example Usage

```terraform
data "aws_spot_datafeed_subscription" "default" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `bucket` - The name of the Amazon S3 bucket where the spot instance data feed is located.
* `prefix` - The prefix for the data feed files.
