---
subcategory: "Route 53 Profiles"
layout: "aws"
page_title: "AWS: aws_route53profiles_profile"
description: |-
  Provides details about an AWS Route 53 Profile.
---

# Data Source: aws_route53profiles_profile

Provides details about an AWS Route 53 Profile.

## Example Usage

### By Name

```terraform
data "aws_route53profiles_profile" "example" {
  name = "example"
}
```

### By ID

```terraform
data "aws_route53profiles_profile" "example" {
  id = "rp-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) ID of the Profile. One of `name` or `id` must be specified.
* `name` - (Optional) Name of the Profile. One of `name` or `id` must be specified.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Profile.
* `id` - ID of the Profile.
* `name` - Name of the Profile.
* `owner_id` - ID of the AWS account that owns the Profile.
* `share_status` - Share status of the Profile.
* `status` - Status of the Profile.
* `status_message` - Status message of the Profile.
* `tags` - Map of tags assigned to the Profile.
