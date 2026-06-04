---
subcategory: "Interconnect"
layout: "aws"
page_title: "AWS: aws_interconnect_connections"
description: |-
  Terraform data source for listing AWS Interconnect Connections.
---

# Data Source: aws_interconnect_connections

Terraform data source for listing AWS Interconnect Connections.

## Example Usage

### Basic Usage

```terraform
data "aws_interconnect_connections" "example" {}
```

### Filter by Environment

```terraform
data "aws_interconnect_connections" "example" {
  environment_id = "example-environment-id"
}
```

## Argument Reference

The following arguments are optional:

* `environment_id` - (Optional) Identifier of the Environment used to filter the connections.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `connections` - List of connections. [See below](#connections).

### connections

* `arn` - ARN of the connection.
* `attach_point` - Attach point to which the connection logically connects.
* `bandwidth` - Bandwidth of the connection.
* `billing_tier` - Billing tier this connection is currently assigned.
* `description` - Description of the connection.
* `environment_id` - Environment on which the connection is placed.
* `id` - Identifier of the connection.
* `interconnect_provider` - Name of the provider on the remote side of this connection.
* `location` - Provider-specific location on the remote side of this connection.
* `shared_id` - Identifier used by both AWS and the remote partner to identify the connection.
* `state` - State of the connection.
* `type` - Specific product type of this connection.
