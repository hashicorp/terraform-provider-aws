---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_configuration_set"
description: |-
  Manages an AWS End User Messaging SMS Configuration Set.
---

# Resource: aws_pinpointsmsvoicev2_configuration_set

Manages an AWS End User Messaging SMS Configuration Set.

## Example Usage

```terraform
resource "aws_pinpointsmsvoicev2_configuration_set" "example" {
  name                 = "example-configuration-set"
  default_sender_id    = "example"
  default_message_type = "TRANSACTIONAL"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the configuration set.
* `default_sender_id` - (Optional) The default sender ID to use for this configuration set.
* `default_message_type` - (Optional) The default message type. Must either be "TRANSACTIONAL" or "PROMOTIONAL"
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the configuration set.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import -out lists using the `name`. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_configuration_set.example
  id = "example-configuration-set"
}
```

Using `terraform import`, import configuration sets using the `name`. For example:

```console
% terraform import aws_pinpointsmsvoicev2_configuration_set.example example-configuration-set
```
