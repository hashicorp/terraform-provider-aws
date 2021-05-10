---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_member"
description: |-
  Provides a resource to manage an Amazon Macie Member.
---

# Resource: aws_macie2_member

Provides a resource to manage an [Amazon Macie Member](https://docs.aws.amazon.com/macie/latest/APIReference/members-id.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {}


resource "aws_macie2_member" "test" {
  account_id = "NAME OF THE MEMBER"
  email      = "EMAIL"
  depends_on = [aws_macie2_account.test]
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) The AWS account ID for the account.
* `email` - (Required) The email address for the account.
* `tags` - (Optional) A map of key-value pairs that specifies the tags to associate with the account in Amazon Macie.
* `status` - (Optional) Specifies the status for the account. To enable Amazon Macie and start all Macie activities for the account, set this value to `ENABLED`. Valid values are `ENABLED` or `PAUSED`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie Member.
* `arn` - The Amazon Resource Name (ARN) of the account.
* `relationship_status` - The current status of the relationship between the account and the administrator account.
* `administrator_account_id` - The AWS account ID for the administrator account.
* `invited_at` - The date and time, in UTC and extended RFC 3339 format, when an Amazon Macie membership invitation was last sent to the account. This value is null if a Macie invitation hasn't been sent to the account.
* `updated_at` - The date and time, in UTC and extended RFC 3339 format, of the most recent change to the status of the relationship between the account and the administrator account.

## Import

`aws_macie2_member` can be imported using the id, e.g.

```
$ terraform import aws_macie2_member.example abcd1
```
