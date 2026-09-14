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

* `listener_identifier` - (Required) ID or ARN of the listener
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_identifier` - (Required) ID or ARN of the service network

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the listener.
* `created_at` - Date and time that the listener was created.
* `default_action` - Action for the default listener rule. See [`default_action` Block](#default_action-block) below.
* `last_updated_at` - Date and time the listener was last updated.
* `listener_id` - ID of the listener.
* `name` - Name of the listener.
* `port` - Listener port.
* `protocol` - Listener protocol. Either `HTTPS` or `HTTP`.
* `service_arn` - ARN of the service.
* `service_id` - ID of the service.
* `tags` - List of tags associated with the listener.

### `default_action` Block

* `fixed_response` - Fixed response action. See [`fixed_response` Block](#fixed_response-block) below.
* `forward` - Forward action. See [`forward` Block](#forward-block) below.

### `fixed_response` Block

* `status_code` - Custom HTTP status code to return.

### `forward` Block

* `target_groups` - Target groups that the listener forwards traffic to. See [`target_groups` Block](#target_groups-block) below.

### `target_groups` Block

* `target_group_identifier` - ID or ARN of the target group.
* `weight` - Weight assigned to the target group that determines the proportion of traffic it receives.
