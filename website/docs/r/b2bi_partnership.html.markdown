---
subcategory: "B2B Data Interchange"
layout: "aws"
page_title: "AWS: aws_b2bi_partnership"
description: |-
  Manages a B2BI Partnership.
---

# Resource: aws_b2bi_partnership

Manages an AWS B2B Data Interchange Partnership. A partnership represents the connection between you and your trading partner, tying together a profile and one or more trading capabilities.

## Example Usage

```terraform
resource "aws_b2bi_partnership" "example" {
  name         = "example-partnership"
  email        = "partner@example.com"
  phone        = "5555555555"
  profile_id   = aws_b2bi_profile.example.profile_id
  capabilities = [aws_b2bi_capability.example.capability_id]
}
```

## Argument Reference

* `capabilities` - (Required) List of capability IDs associated with this partnership.
* `email` - (Required, Forces new resource) The email address of the trading partner.
* `name` - (Required) The name of the partnership.
* `profile_id` - (Required, Forces new resource) The ID of the profile to associate with this partnership.
* `phone` - (Optional, Forces new resource) The phone number of the trading partner.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attribute Reference

* `partnership_arn` - The ARN of the partnership.
* `partnership_id` - The unique identifier of the partnership.
* `trading_partner_id` - The ID of the trading partner.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

B2BI Partnerships can be imported using the `partnership_id`:

```console
% terraform import aws_b2bi_partnership.example ps-5555zzzz4444yyyyy
```
