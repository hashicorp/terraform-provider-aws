---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_storage_lens_configuration"
description: |-
  Provides a resource to manage an S3 Storage Lens configuration.
---

# Resource: aws_s3control_storage_lens_configuration

Provides a resource to manage an S3 Storage Lens configuration.

## Example Usage

```terraform
resource "aws_s3control_storage_lens_configuration" "example" {
  config_id = "example-1"

  storage_lens_configuration {
  
  }
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) The AWS account ID for the S3 Storage Lens configuration. Defaults to automatically determined account ID of the Terraform AWS provider.
* `config_id` - (Required) The ID of the S3 Storage Lens configuration.
* `storage_lens_configuration` - (Required) The S3 Storage Lens configuration. See [Storage Lens Configuration](#storage-lens-configuration) below for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Storage Lens Configuration

The `storage_lens_configuration` block supports the following:

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the S3 Storage Lens configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

S3 Storage Lens configurations can be imported using the `account_id` and `config_id`, separated by a colon (`:`), e.g.

```
$ terraform import aws_s3control_storage_lens_configuration.example 123456789012:example-1
```