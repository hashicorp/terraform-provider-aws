---
subcategory: "B2B Data Interchange"
layout: "aws"
page_title: "AWS: aws_b2bi_profile"
description: |-
  Manages a B2BI Profile.
---

# Resource: aws_b2bi_profile

Manages an AWS B2B Data Interchange Profile. A profile is the mechanism used to create the concept of a private network. You can have up to five customer profiles, each representing a distinct private network.

See the [B2B Data Interchange User Guide](https://docs.aws.amazon.com/b2bi/latest/userguide/what-is-b2bi.html) for more information.

## Example Usage

```terraform
resource "aws_b2bi_profile" "example" {
  name          = "example-profile"
  business_name = "Example Corp"
  phone         = "5555555555"
  email         = "admin@example.com"
  logging       = "ENABLED"
}
```

## Argument Reference

This resource supports the following arguments:

* `business_name` - (Required) The name of the business associated with this profile.
* `logging` - (Required, Forces new resource) Whether logging is enabled for this profile. Valid values are `ENABLED` and `DISABLED`.
* `name` - (Required) The name of the profile.
* `phone` - (Required) The phone number associated with the profile.
* `email` - (Optional) The email address associated with this customer profile.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `log_group_name` - The name of the logging group.
* `profile_arn` - The Amazon Resource Name (ARN) of the profile.
* `profile_id` - The unique identifier of the profile.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import B2BI Profiles using the `profile_id`. For example:

```terraform
import {
  to = aws_b2bi_profile.example
  id = "p-ABCDE111122223333"
}
```

Using `terraform import`, import B2BI Profiles using the `profile_id`. For example:

```console
% terraform import aws_b2bi_profile.example p-ABCDE111122223333
```
