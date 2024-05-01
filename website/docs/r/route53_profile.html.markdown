---
subcategory: "Route53"
layout: "aws"
page_title: "AWS: aws_route53profiles_profile"
description: |-
  Manages Route53 Profile
---

# Resource: aws_route53profiles_profile

Manages Route53 Profile

## Example Usage

### Basic Usage

```terraform
resource "aws_route53profiles_profile" "example" {
    name = "example"
}

output "profile_id" {
    value = aws_route53profiles_profile.example.id
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the profile (up to 64 letters, numbers, hyphens, and underscores)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the Profile.
* `arn` - The Amazon Resource Name (ARN) of the Profile.
* `client_token` - The `ClientToken` value that was assigned when the Profile was created.