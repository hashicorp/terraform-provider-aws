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

This data source supports the following arguments:

* `listener_identifier` - (Required) ID or Amazon Resource Name (ARN) of the listener
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_identifier` - (Required) ID or Amazon Resource Name (ARN) of the service network

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the listener.
* `created_at` - Date and time that the listener was created.
* `default_action` - Actions for the default listener rule.
* `last_updated_at` - Date and time the listener was last updated.
* `listener_id` - ID of the listener.
* `name` - Name of the listener.
* `port` - Listener port.
* `protocol` - Listener protocol. Either `HTTPS` or `HTTP`.
* `service_arn` - ARN of the service.
* `service_id` - ID of the service.
* `tags` - List of tags associated with the listener.
