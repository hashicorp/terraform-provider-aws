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

* `concurrent_jobs` - (Optional) Maximum number of jobs your queue can process concurrently. For on-demand queues, the value you enter is constrained by your service quotas for Maximum concurrent jobs, per on-demand queue and Maximum concurrent jobs, per account. For reserved queues, specify the number of jobs you can process concurrently in your reservation plan instead.
* `description` - (Optional) Description of the queue
* `name` - (Required) Unique identifier describing the queue
* `pricing_plan` - (Optional) Whether the pricing plan for the queue is on-demand or reserved. Valid values are `ON_DEMAND` or `RESERVED`. Default to `ON_DEMAND`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `reservation_plan_settings` - (Optional) Detail pricing plan of the reserved queue. See below.
* `status` - (Optional) Status of the queue. Valid values are `ACTIVE` or `RESERVED`. Default to `PAUSED`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `reservation_plan_settings` Block

* `commitment` - (Required) Length of the term of your reserved queue pricing plan commitment. Valid value is `ONE_YEAR`.
* `renewal_type` - (Required) Renewal type of your reserved queue pricing plan. Valid values are `AUTO_RENEW` or `EXPIRE`.
* `reserved_slots` - (Required) Number of reserved transcode slots (RTS) for queue.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the queue
* `id` - Same as `name`
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
