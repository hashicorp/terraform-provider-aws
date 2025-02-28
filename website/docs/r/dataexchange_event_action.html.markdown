---
subcategory: "Data Exchange"
layout: "aws"
page_title: "AWS: aws_dataexchange_event_action"
description: |-
  Terraform resource for managing an AWS Data Exchange Event Action.
---

# Resource: aws_dataexchange_event_action

Terraform resource for managing an AWS Data Exchange Event Action.

## Example Usage

```terraform
resource "aws_dataexchange_event_action" "example" {
  event {
    revision_published {
      data_set_id = aws_dataexchange_data_set.example.id
    }
  }

  action {
    export_revision_to_s3 {
      revision_destination {
        bucket      = aws_s3_bucket.example.bucket
        key_pattern = "$${Revision.CreatedAt}/$${Asset.Name}"
      }

      encryption {
        type        = "aws:kms"
        kms_key_arn = aws_kms_key.example.arn
      }
    }
  }
}
```

## Argument Reference

The following blocks are supported:

* `action` - (Required) Describes the action to take.
  Described in [`action` Configuration Block](#action-configuration-block) below.
* `event` - (Required) Describes the event that triggers the `action`.
  Described in [`event` Configuration Block](#event-configuration-block) below.

### action Configuration Block

* `export_revision_to_s3` - (Required) Configuration for an Export Revision to S3 action.
  Described in [`export_revision_to_s3` Configuration Block](#export_revision_to_s3-configuration-block)

### export_revision_to_s3 Configuration Block

* `encryption` - (Optional) Configures server-side encryption of the exported revision.
  Described in [`encryption` Configuration Block](#encryption-configuration-block) below.
* `revision_destination` - (Required) Configures the S3 destination of the exported revision.
  Described in [`revision_destination` Configuration Block](#revision_destination-configuration-block) below.

### encryption Configuration Block

* `type` - (Optional) Type of server-side encryption.
  Valid values are `aws:kms` or `aws:s3`.
* `kms_key_arn` - (Optional) ARN of the KMS key used for encryption.

### revision_destination Configuration Block

* `bucket` - (Required) The S3 bucket where the revision will be exported.
* `key_pattern` - (Optional) Pattern for naming revisions in the S3 bucket.
  Defaults to `${Revision.CreatedAt}/${Asset.Name}`.

### event Configuration Block

* `revision_published` - (Required) Configuration for a Revision Published event.
  Described in [`revision_published` Configuration Block](#revision_published-configuration-block) below.

### revision_published Configuration Block

* `data_set_id` - (Required) The ID of the data set to monitor for revision publications.
  Changing this value will recreate the resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the event action.
* `created_at` - Date and time when the resource was created.
* `id` - Unique identifier for the event action.
* `updated_at` - Data and time when the resource was last updated.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Data Exchange Event Action using the `id`. For example:

```terraform
import {
  to = aws_dataexchange_event_action.example
  id = "example-event-action-id"
}
```

Using `terraform import`, import Data Exchange Event Action using the id. For example:

```console
% terraform import aws_dataexchange_event_action.example example-event-action-id
```
