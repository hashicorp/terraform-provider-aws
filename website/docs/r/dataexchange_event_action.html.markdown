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
  event_revision_published {
    data_set_id = aws_dataexchange_data_set.example.id
  }

  action_export_revision_to_s3 {
    revision_destination {
      bucket = aws_s3_bucket.example.id
      key_pattern = "\${Revision.CreatedAt}/\${Asset.Name}"
    }
    
    encryption {
      type = "aws:kms"
      kms_key_arn = aws_kms_key.example.arn
    }
  }
}
```

## Argument Reference

The following blocks are supported:

* `event_revision_published` - (Required) Configuration block for the revision published event that triggers the action. Cannot be updated without recreation of the resource.
* `action_export_revision_to_s3` - (Required) Configuration block for the export revision to S3 action.

### event_revision_published Configuration Block

* `data_set_id` - (Required) The ID of the data set to monitor for revision publications.

### action_export_revision_to_s3 Configuration Block

* `revision_destination` - (Required) Configuration block for the S3 destination of the exported revision.
* `encryption` - (Optional) Configuration block for server-side encryption of the exported revision.

### revision_destination Configuration Block

* `bucket` - (Required) The S3 bucket where the revision will be exported.
* `key_pattern` - (Optional) Pattern for naming revisions in the S3 bucket. Defaults to "\${Revision.CreatedAt}/\${Asset.Name}".

### encryption Configuration Block

* `type` - (Optional) Type of server-side encryption. Valid values are `aws:kms` or `aws:s3`.
* `kms_key_arn` - (Optional) ARN of the KMS key used for encryption.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the event action.
* `id` - Unique identifier for the event action.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

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