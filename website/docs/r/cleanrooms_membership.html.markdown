---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_membership"
description: |-
  Provides a Clean Rooms Membership.
---

# Resource: aws_cleanrooms_membership

Provides a AWS Clean Rooms membership. Memberships are used to join a Clean Rooms collaboration by the invited member.

## Example Usage

### Membership with tags

```terraform
resource "aws_cleanrooms_membership" "test_membership" {
  collaboration_id = "1234abcd-12ab-34cd-56ef-1234567890ab"
  query_log_status = "DISABLED"

  default_result_configuration {
    role_arn = "arn:aws:iam::123456789012:role/role-name"
    output_configuration {
      s3 {
        bucket        = "test-bucket"
        result_format = "PARQUET"
        key_prefix    = "test-prefix"
      }
    }
  }

  tags = {
    Project = "Terraform"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `collaboration_id` - (Required - Forces new resource) - The ID of the collaboration to which the member was invited.
* `query_log_status` - (Required) - An indicator as to whether query logging has been enabled or disabled for the membership.
* `default_result_configuration` - (Optional) - The default configuration for a query result.
    - `role_arn` - (Optional) - The ARN of the IAM role which will be used to create the membership.
    - `output_configuration.s3.bucket` - (Required) - The name of the S3 bucket where the query results will be stored.
    - `output_configuration.s3.result_format` - (Required) - The format of the query results. Valid values are `PARQUET` and `CSV`.
    - `output_configuration.s3.key_prefix` - (Optional) - The prefix used for the query results.
* `tags` - (Optional) - Key value pairs which tag the membership.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the membership.
* `collaboration_arn` - The ARN of the joined collaboration.
* `collaboration_creator_account_id` - The account ID of the collaboration's creator.
* `collaboration_creator_display_name` - The display name of the collaboration's creator.
* `collaboration_id` - The ID of the joined collaboration.
* `collaboration_name` - The name of the joined collaboration.
* `create_time` - The date and time the membership was created.
* `id` - The ID of the membership.
* `member_abilities` - The list of abilities for the invited member.
* `payment_configuration.query_compute.is_responsible` - Indicates whether the collaboration member has accepted to pay for query compute costs.
* `status` - The status of the membership.
* `update_time` - The date and time the membership was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `1m`)
- `update` - (Default `1m`)
- `delete` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_cleanrooms_membership` using the `id`. For example:

```terraform
import {
  to = aws_cleanrooms_membership.membership
  id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

Using `terraform import`, import `aws_cleanrooms_membership` using the `id`. For example:

```console
% terraform import aws_cleanrooms_membership.membership 1234abcd-12ab-34cd-56ef-1234567890ab
```
