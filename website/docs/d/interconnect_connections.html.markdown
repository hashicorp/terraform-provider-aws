---
subcategory: "Interconnect"
layout: "aws"
page_title: "AWS: aws_interconnect_connections"
description: |-
  Terraform data source for listing AWS Interconnect Connections.
---

# Data Source: aws_interconnect_connections

Terraform data source for listing the AWS Interconnect Connections to which the caller has access.

## Example Usage

### Basic Usage

```terraform
data "aws_interconnect_connections" "example" {}
```

### Filtering

Filters are combined, so only Connections matching every filter supplied are returned.

```terraform
data "aws_interconnect_connections" "example" {
  environment_id = "mce-aws-gcp-iad"
  state          = "available"

  interconnect_provider {
    cloud_service_provider = "gcp"
  }
}
```

## Argument Reference

The following arguments are optional:

* `attach_point` - (Optional) Filters results to Connections attached to the given attach point. See [`attach_point`](#attach_point-block) below for details.
* `environment_id` - (Optional) Filters results to Connections on the given Environment.
* `interconnect_provider` - (Optional) Filters results to Connections to the given provider on the remote side. See [`interconnect_provider`](#interconnect_provider-block) below for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `state` - (Optional) Filters results to Connections in the given state. One of `requested`, `pending`, `available`, `down`, `deleting`, `deleted`, `failed`, or `updating`.

### `attach_point` Block

Exactly one of the following arguments must be set.

* `arn` - (Optional) ARN of the attach point.
* `direct_connect_gateway` - (Optional) Identifier of a Direct Connect Gateway attach point.

### `interconnect_provider` Block

Exactly one of the following arguments must be set.

* `cloud_service_provider` - (Optional) Name of a cloud service provider. Connections to this provider are considered Multicloud connections.
* `last_mile_provider` - (Optional) Name of a last mile provider. Connections to this provider are considered LastMile connections.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `connections` - List of connections. See [`connections`](#connections-block) below for details.

### `connections` Block

* `arn` - ARN of the connection.
* `attach_point` - Attach point to which the connection logically connects within your AWS network. See [`connections` `attach_point`](#connections-attach_point-block) below for details.
* `bandwidth` - Bandwidth of the connection.
* `billing_tier` - Billing tier this connection is currently assigned.
* `description` - Description of the connection.
* `environment_id` - Identifier of the Environment on which the connection was created.
* `id` - Identifier of the connection.
* `interconnect_provider` - Name of the provider on the remote side of the connection.
* `location` - Provider-specific location on the remote side of the connection.
* `shared_id` - Identifier used by both AWS and the remote partner to identify the connection.
* `state` - State of the connection.
* `type` - Specific product type of the connection.

### `connections` `attach_point` Block

* `arn` - ARN of the attach point.
* `direct_connect_gateway` - Identifier of the Direct Connect Gateway attach point.
