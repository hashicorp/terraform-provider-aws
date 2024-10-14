---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_user_profile"
description: |-
  Terraform resource for managing an AWS DataZone User Profile.
---

# Resource: aws_datazone_user_profile

Terraform resource for managing an AWS DataZone User Profile.

## Example Usage

### Basic Usage

```terraform
resource "aws_datazone_user_profile" "example" {
  user_identifier   = aws_iam_user.example.arn
  domain_identifier = aws_datazone_domain.example.id
  user_type         = "IAM_USER"
}
```

## Argument Reference

The following arguments are required:

* `domain_identifier` - (Required) The domain identifier.
* `user_identifier` - (Required) The user identifier.

The following arguments are optional:

* `status` - (Optional) The user profile status.
* `user_type` - (Optional) The user type.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `details` - Details about the user profile.
* `id` - The user profile identifier.
* `type` - The user profile type.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone User Profile using the `user_identifier,domain_identifier,type`. For example:

```terraform
import {
  to = aws_datazone_user_profile.example
  id = "arn:aws:iam::012345678901:user/example,dzd_54nakfrg9k6sri,IAM"
}
```

Using `terraform import`, import DataZone User Profile using the `user_identifier,domain_identifier,type`. For example:

```console
% terraform import aws_datazone_user_profile.example arn:aws:iam::012345678901:user/example,dzd_54nakfrg9k6suo,IAM
```
