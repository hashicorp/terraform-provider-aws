---
subcategory: "Route 53 Profiles"
layout: "aws"
page_title: "AWS: aws_route53profiles_profiles"
description: |-
  Terraform data source for managing an AWS Route 53 Profiles.
---

# Data Source: aws_route53profiles_profiles

Terraform data source for managing an AWS Route 53 Profiles.

## Example Usage

### Basic Usage

```terraform
data "aws_route53profiles_profiles" "example" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

This data source exports the following attributes:

* `profiles` - List of Profiles.
    * `arn` - The ARN of the Profile.
    * `id` - The ID of the Profile.
    * `name` - The name of the Profile.
    * `share_status` - The share status of the Profile. Valid values [AWS docs](https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53profiles_Profile.html)
