---
subcategory: "Elemental MediaConvert"
layout: "aws"
page_title: "AWS: aws_media_convert_queue"
description: |-
  Provides an AWS Elemental MediaConvert Queue.
---

# Resource: aws_media_convert_queue

Provides an AWS Elemental MediaConvert Queue.

## Example Usage

```terraform
resource "aws_media_convert_queue" "test" {
  name = "tf-test-queue"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A unique identifier describing the queue
* `description` - (Optional) A description of the queue
* `pricing_plan` - (Optional) Specifies whether the pricing plan for the queue is on-demand or reserved. Valid values are `ON_DEMAND` or `RESERVED`. Default to `ON_DEMAND`.
* `reservation_plan_settings` - (Optional) A detail pricing plan of the  reserved queue. See below.
* `status` - (Optional) A status of the queue. Valid values are `ACTIVE` or `RESERVED`. Default to `PAUSED`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Nested Fields

#### `reservation_plan_settings`

* `commitment` - (Required) The length of the term of your reserved queue pricing plan commitment. Valid value is `ONE_YEAR`.
* `renewal_type` - (Required) Specifies whether the term of your reserved queue pricing plan. Valid values are `AUTO_RENEW` or `EXPIRE`.
* `reserved_slots` - (Required) Specifies the number of reserved transcode slots (RTS) for queue.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The same as `name`
* `arn` - The Arn of the queue
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Media Convert Queue using the queue name. For example:

```terraform
import {
  to = aws_media_convert_queue.test
  id = "tf-test-queue"
}
```

Using `terraform import`, import Media Convert Queue using the queue name. For example:

```console
% terraform import aws_media_convert_queue.test tf-test-queue
```
