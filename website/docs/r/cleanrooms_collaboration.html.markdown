---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_collaboration"
description: |-
  Provides a Clean Rooms Collaboration.
---

# Resource: aws_cleanrooms_collaboration

Provides a AWS Clean Rooms collaboration.
All members included in the definition will be invited to join the collaboration and can create memberships.

## Example Usage

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

The following arguments are required:

* `creator_display_name` - (Required - Forces new resource) - Name for the member record for the collaboration creator.
* `creator_member_abilities` - (Required - Forces new resource) - List of member abilities for the creator of the collaboration. Valid values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_CreateCollaboration.html#API-CreateCollaboration-request-creatorMemberAbilities).
* `description` - (Required) - Description for a collaboration.
* `name` - (Required) - Name of the collaboration.  Collaboration names do not need to be unique.
* `query_log_status` - (Required - Forces new resource) Whether members of the collaboration can enable query logs within their own memberships. Valid values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_CreateCollaboration.html#API-CreateCollaboration-request-queryLogStatus).

The following arguments are optional:

* `allowed_result_regions` - (Optional - Forces new resource) AWS Regions where collaboration query results can be stored. Valid values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_SupportedS3Region.html).
* `analytics_engine` - (Optional - Forces new resource, **deprecated**) Analytics engine for the collaboration. Spark is now the only engine accepted by AWS for new collaborations; supplying `CLEAN_ROOMS_SQL` results in a `ValidationException` at apply time. Omitting this argument lets AWS apply its default. See the [AWS Clean Rooms document history](https://docs.aws.amazon.com/clean-rooms/latest/userguide/doc-history.html).
* `auto_approved_change_request_types` - (Optional - Forces new resource) Types of change requests that are automatically approved for this collaboration. Valid values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_AutoApprovedChangeType.html).
* `creator_ml_member_abilities` - (Optional - Forces new resource) ML abilities granted to the collaboration creator. [See below](#ml_member_abilities-configuration-block).
* `creator_payment_configuration` - (Optional - Forces new resource) Collaboration creator's payment responsibilities for query, job, and ML compute costs. [See below](#payment_configuration-configuration-block).
* `data_encryption_metadata` - (Optional - Forces new resource) Collection of settings which determine how the [c3r client](https://docs.aws.amazon.com/clean-rooms/latest/userguide/crypto-computing.html) will encrypt data for use within this collaboration. [See below](#data_encryption_metadata-configuration-block).
* `is_metrics_enabled` - (Optional - Forces new resource) Whether collaboration members can opt in to Amazon CloudWatch metrics for their membership queries.
* `job_log_status` - (Optional - Forces new resource) Whether job logs are enabled for this collaboration. Valid values are `ENABLED` and `DISABLED`.
* `member` - (Optional - Forces new resource) Additional members of the collaboration which will be invited to join the collaboration. [See below](#member-configuration-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key value pairs which tag the collaboration.

### `data_encryption_metadata` Configuration Block

* `allow_clear_text` - (Required - Forces new resource) Whether encrypted tables can contain cleartext data. This is a boolean field.
* `allow_duplicates` - (Required - Forces new resource) Whether Fingerprint columns can contain duplicate entries. This is a boolean field.
* `allow_joins_on_columns_with_different_names` - (Required - Forces new resource) Whether Fingerprint columns can be joined on any other Fingerprint column with a different name. This is a boolean field.
* `preserve_nulls` - (Required - Forces new resource) Whether NULL values are to be copied as NULL to encrypted tables (true) or cryptographically processed (false).

### `member` Configuration Block

* `account_id` - (Required - Forces new resource) Account ID for the invited member.
* `display_name` - (Required - Forces new resource) Display name for the invited member.
* `member_abilities` - (Required - Forces new resource) List of abilities for the invited member. Valid values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_CreateCollaboration.html#API-CreateCollaboration-request-creatorMemberAbilities).
* `ml_member_abilities` - (Optional - Forces new resource) ML abilities granted to the invited member. [See below](#ml_member_abilities-configuration-block).
* `payment_configuration` - (Optional - Forces new resource) Invited member's payment responsibilities for query, job, and ML compute costs. [See below](#payment_configuration-configuration-block).

### `ml_member_abilities` Configuration Block

* `custom_ml_member_abilities` - (Required - Forces new resource) Custom ML abilities granted to the member. Valid values are `CAN_RECEIVE_MODEL_OUTPUT` and `CAN_RECEIVE_INFERENCE_OUTPUT`.

### `payment_configuration` Configuration Block

* `query_compute` - (Required - Forces new resource) Payment responsibilities for query compute costs. [See below](#is_responsible-configuration-block).
* `job_compute` - (Optional - Forces new resource) Payment responsibilities for job compute costs. [See below](#is_responsible-configuration-block).
* `machine_learning` - (Optional - Forces new resource) Payment responsibilities for ML compute costs. [See below](#machine_learning-configuration-block).

#### `machine_learning` Configuration Block

* `model_inference` - (Optional - Forces new resource) Payment responsibilities for model inference. [See below](#is_responsible-configuration-block).
* `model_training` - (Optional - Forces new resource) Payment responsibilities for model training. [See below](#is_responsible-configuration-block).
* `synthetic_data_generation` - (Optional - Forces new resource) Payment responsibilities for synthetic data generation. [See below](#is_responsible-configuration-block).

#### `is_responsible` Configuration Block

The `query_compute`, `job_compute`, `model_inference`, `model_training`, and `synthetic_data_generation` blocks all share this shape:

* `is_responsible` - (Required - Forces new resource) Whether the member is responsible for the corresponding compute costs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the collaboration.
* `create_time` - Date and time the collaboration was created.
* `id` - ID of the collaboration.
* `member.status` - For each member included in the collaboration an additional computed attribute of status is added. These values [may be found here](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_MemberSummary.html#API-Type-MemberSummary-status).
* `membership_arn` - The unique ARN for the calling account's membership within the collaboration, if present. May be empty when no associated membership has been created with `aws_cleanrooms_membership`. See [`API_Collaboration#membershipArn`](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_Collaboration.html).
* `membership_id` - The unique ID for the calling account's membership within the collaboration, if present. May be empty when no associated membership has been created with `aws_cleanrooms_membership`. See [`API_Collaboration#membershipId`](https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_Collaboration.html).
* `update_time` - Date and time the collaboration was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `1m`)
- `update` - (Default `1m`)
- `delete` - (Default `1m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cleanrooms_collaboration.collaboration
  identity = {
    id = "1234abcd-12ab-34cd-56ef-1234567890ab"
  }
}

resource "aws_cleanrooms_collaboration" "collaboration" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the cleanrooms collaboration.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

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
