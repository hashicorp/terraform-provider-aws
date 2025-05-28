---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_collaboration"
description: |-
  Provides a Clean Rooms Collaboration.
---

# Resource: aws_cleanrooms_collaboration

Provides a AWS Clean Rooms collaboration.  All members included in the definition will be invited to
join the collaboration and can create memberships.

## Example Usage

### Collaboration with tags

```terraform
resource "aws_cleanrooms_collaboration" "test_collaboration" {
  name                     = "terraform-example-collaboration"
  creator_member_abilities = ["CAN_QUERY", "CAN_RECEIVE_RESULTS"]
  creator_display_name     = "Creator "
  description              = "I made this collaboration with terraform!"
  query_log_status         = "DISABLED"

  data_encryption_metadata {
    allow_clear_text                            = true
    allow_duplicates                            = true
    allow_joins_on_columns_with_different_names = true
    preserve_nulls                              = false
  }

  member {
    account_id       = 123456789012
    display_name     = "Other member"
    member_abilities = []
  }

  tags = {
    Project = "Terraform"
  }

}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) - The name of the collaboration.  Collaboration names do not need to be unique.
* `description` - (Required) - A description for a collaboration.
* `creator_member_abilities` - (Required - Forces new resource) - The list of member abilities for the creator of the collaboration.  Valid values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_CreateCollaboration.html#API-CreateCollaboration-request-creatorMemberAbilities).
* `creator_display_name` - (Required - Forces new resource) - The name for the member record for the collaboration creator.
* `query_log_status` - (Required - Forces new resource) - Determines if members of the collaboration can enable query logs within their own.
emberships. Valid values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_CreateCollaboration.html#API-CreateCollaboration-request-queryLogStatus).
* `data_encryption_metadata` - (Required - Forces new resource) - a collection of settings which determine how the [c3r client](https://docs.aws.amazon.com/clean-rooms/latest/userguide/crypto-computing.html) will encrypt data for use within this collaboration.
* `data_encryption_metadata.allow_clear_text` - (Required - Forces new resource) - Indicates whether encrypted tables can contain cleartext data. This is a boolea
 field.
* `data_encryption_metadata.allow_duplicates` - (Required - Forces new resource ) - Indicates whether Fingerprint columns can contain duplicate entries. This is a
boolean field.
* `data_encryption_metadata.allow_joins_on_columns_with_different_names` - (Required - Forces new resource) - Indicates whether Fingerprint columns can be joined
n any other Fingerprint column with a different name. This is a boolean field.
* `data_encryption_metadata.preserve_nulls` - (Required - Forces new resource) - Indicates whether NULL values are to be copied as NULL to encrypted tables (true)
or cryptographically processed (false).
* `member` - (Optional - Forces new resource) - Additional members of the collaboration which will be invited to join the collaboration.
* `member.account_id` - (Required - Forces new resource) - The account id for the invited member.
* `member.display_name` - (Required - Forces new resource) - The display name for the invited member.
* `member.member_abilities` - (Required - Forces new resource) - The list of abilities for the invited member. Valid values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_CreateCollaboration.html#API-CreateCollaboration-request-creatorMemberAbilities).
* `tags` - (Optional) - Key value pairs which tag the collaboration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The arn of the collaboration.
* `id` - The id of the collaboration.
* `create_time` - The date and time the collaboration was created.
* `member status` - For each member included in the collaboration an additional computed attribute of status is added. These values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_MemberSummary.html#API-Type-MemberSummary-status).
* `updated_time` - The date and time the collaboration was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `1m`)
- `update` - (Default `1m`)
- `delete` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_cleanrooms_collaboration` using the `id`. For example:

```terraform
import {
  to = aws_cleanrooms_collaboration.collaboration
  id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

Using `terraform import`, import `aws_cleanrooms_collaboration` using the `id`. For example:

```console
% terraform import aws_cleanrooms_collaboration.collaboration 1234abcd-12ab-34cd-56ef-1234567890ab
```
