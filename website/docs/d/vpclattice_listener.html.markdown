---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_listener"
description: |-
  Terraform data source for managing an AWS VPC Lattice Listener.
---

# Data Source: aws_vpclattice_listener

Terraform data source for managing an AWS VPC Lattice Listener.

## Example Usage

### Basic Usage

```terraform
data "aws_vpclattice_listener" "example" {
}
```

## Argument Reference

The following arguments are required:

* `service_identifier` - (Required) ID or Amazon Resource Name (ARN) of the service network
* `listener_identifier` - (Required) ID or Amazon Resource Name (ARN) of the listener

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the listener.
* `created_at` - The date and time that the listener was created.
* `default_action` - The actions for the default listener rule.
* `last_updated_at` - The date and time the listener was last updated.
* `listener_id` - The ID of the listener.
* `name` - The name of the listener.
* `port` - The listener port.
* `protocol` - The listener protocol. Either `HTTPS` or `HTTP`.
* `service_arn` - The ARN of the service.
* `service_id` - The ID of the service.
* `tags` - List of tags associated with the listener.
