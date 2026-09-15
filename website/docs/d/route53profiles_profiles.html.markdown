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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `profiles` - List of Profiles.
    * `arn` - ARN of the Profile.
    * `id` - ID of the Profile.
    * `name` - Name of the Profile.
    * `share_status` - Share status of the Profile. Valid values [AWS docs](https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53profiles_Profile.html)
