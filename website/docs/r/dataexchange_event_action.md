---
subcategory: "DataExchange"
layout: "aws"
page_title: "AWS: aws_dataexchange_event_action"
description: |-
  Provides a DataExchange Event Action resource.
---

# Resource: aws_dataexchange_event_action

Provides a resource to manage AWS DataExchange Event Actions. Event actions allow you to configure automatic responses to events in your data sets, such as automatically exporting new revisions to S3 when they are published.

## Example Usage

```terraform
resource "aws_dataexchange_event_action" "example" {
  action_export_revision_to_s3 {
    bucket      = "my-export-bucket"
    key_pattern = "${Revision.CreatedAt}/${Asset.Name}"
    
    s3_encryption_type     = "aws:kms"
    s3_encryption_kms_key_arn = aws_kms_key.example.arn
  }

  event_revision_published {
    data_set_id = aws_dataexchange_data_set.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `action_export_revision_to_s3` - (Required) Configuration block for the export revision to S3 action. Detailed below.
* `event_revision_published` - (Required) Configuration block for the revision published event trigger. Detailed below.

### action_export_revision_to_s3 Configuration Block

* `bucket` - (Required) The S3 bucket to export the revision to.
* `key_pattern` - (Optional) The pattern for naming files in the S3 bucket. Defaults to `${Revision.CreatedAt}/${Asset.Name}`.
* `s3_encryption_type` - (Optional) The type of server-side encryption to use. Valid values: `aws:kms`.
* `s3_encryption_kms_key_arn` - (Optional) The ARN of the KMS key to use for encryption.

### event_revision_published Configuration Block

* `data_set_id` - (Required) The ID of the data set to monitor for new revisions.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the event action.
* `arn` - The ARN of the event action.
* `created_at` - The date and time the event action was created.
* `last_updated_time` - The date and time the event action was last updated.

## Import

DataExchange Event Actions can be imported using the `id`, e.g.,

```shell
$ terraform import aws_dataexchange_event_action.example ea-12345678