---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_routing_profile"
description: |-
  Provides details about a specific Amazon Connect Routing Profile.
---

# Resource: aws_connect_routing_profile

Provides an Amazon Connect Routing Profile resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

## Example Usage

```terraform
resource "aws_connect_routing_profile" "example" {
  instance_id               = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name                      = "example"
  default_outbound_queue_id = "12345678-1234-1234-1234-123456789012"
  description               = "example description"

  media_concurrencies {
    channel     = "VOICE"
    concurrency = 1
  }

  queue_configs {
    channel  = "VOICE"
    delay    = 2
    priority = 1
    queue_id = "12345678-1234-1234-1234-123456789012"
  }

  tags = {
    "Name" = "Example Routing Profile",
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `default_outbound_queue_id` - (Required) Specifies the default outbound queue for the Routing Profile.
* `description` - (Required) Specifies the description of the Routing Profile.
* `instance_id` - (Required) Specifies the identifier of the hosting Amazon Connect Instance.
* `media_concurrencies` - (Required) One or more `media_concurrencies` blocks that specify the channels that agents can handle in the Contact Control Panel (CCP) for this Routing Profile. The `media_concurrencies` block is documented below.
* `name` - (Required) Specifies the name of the Routing Profile.
* `queue_configs` - (Optional) One or more `queue_configs` blocks that specify the inbound queues associated with the routing profile. If no queue is added, the agent only can make outbound calls. The `queue_configs` block is documented below.
* `tags` - (Optional) Tags to apply to the Routing Profile. If configured with a provider
[`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

A `media_concurrencies` block supports the following arguments:

* `channel` - (Required) Specifies the channels that agents can handle in the Contact Control Panel (CCP). Valid values are `VOICE`, `CHAT`, `TASK`.
* `concurrency` - (Required) Specifies the number of contacts an agent can have on a channel simultaneously. Valid Range for `VOICE`: Minimum value of 1. Maximum value of 1. Valid Range for `CHAT`: Minimum value of 1. Maximum value of 10. Valid Range for `TASK`: Minimum value of 1. Maximum value of 10.

A `queue_configs` block supports the following arguments:

* `channel` - (Required) Specifies the channels agents can handle in the Contact Control Panel (CCP) for this routing profile. Valid values are `VOICE`, `CHAT`, `TASK`.
* `delay` - (Required) Specifies the delay, in seconds, that a contact should be in the queue before they are routed to an available agent
* `priority` - (Required) Specifies the order in which contacts are to be handled for the queue.
* `queue_id` - (Required) Specifies the identifier for the queue.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the Routing Profile.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the Routing Profile separated by a colon (`:`).
* `queue_configs` - In addition to the arguments used in the `queue_configs` argument block, there are additional attributes exported within the `queue_configs` block. These additional attributes are documented below.
* `routing_profile_id` - The identifier for the Routing Profile.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

A `queue_configs` block supports the following attributes in addition to the arguments defined earlier:

* `queue_arn` - ARN for the queue.
* `queue_name` - Name for the queue.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Connect Routing Profiles using the `instance_id` and `routing_profile_id` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_connect_routing_profile.example
  id = "f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5"
}
```

Using `terraform import`, import Amazon Connect Routing Profiles using the `instance_id` and `routing_profile_id` separated by a colon (`:`). For example:

```console
% terraform import aws_connect_routing_profile.example f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5
```
