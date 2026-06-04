---
subcategory: "Interconnect"
layout: "aws"
page_title: "AWS: aws_interconnect_connection"
description: |-
  Terraform data source for an AWS Interconnect Connection.
---

# Data Source: aws_interconnect_connection

Terraform data source for an AWS Interconnect Connection.

## Example Usage

### Basic Usage

```terraform
data "aws_interconnect_connection" "example" {
  id = "mcc-abcd1234"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Identifier of the connection.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `activation_key` - Activation key associated with this connection.
* `arn` - ARN of the connection.
* `attach_point` - Attach point to which the connection logically connects. [See below](#attach_point).
* `bandwidth` - Bandwidth of the connection.
* `billing_tier` - Billing tier this connection is currently assigned.
* `description` - Description of the connection.
* `environment_id` - Environment on which the connection is placed.
* `interconnect_provider` - Name of the provider on the remote side of this connection.
* `location` - Provider-specific location on the remote side of this connection.
* `owner_account` - Account that owns this connection.
* `shared_id` - Identifier used by both AWS and the remote partner to identify the connection.
* `state` - State of the connection.
* `tags` - Map of tags assigned to the resource.
* `type` - Specific product type of this connection.

### attach_point

* `arn` - ARN of the attach point.
* `direct_connect_gateway` - Identifier of the Direct Connect Gateway attach point.
