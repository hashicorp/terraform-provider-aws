---
subcategory: "Interconnect"
layout: "aws"
page_title: "AWS: aws_interconnect_environments"
description: |-
  Terraform data source for listing AWS Interconnect Environments.
---

# Data Source: aws_interconnect_environments

Terraform data source for listing AWS Interconnect Environments.

## Example Usage

### Basic Usage

```terraform
data "aws_interconnect_environments" "example" {}
```

### Filter by Location

```terraform
data "aws_interconnect_environments" "example" {
  location = "us-east4"
}
```

### Filter by Provider

```terraform
data "aws_interconnect_environments" "example" {
  interconnect_provider {
    cloud_service_provider = "gcp"
  }
}
```

## Argument Reference

The following arguments are optional:

* `interconnect_provider` - (Optional) Filters results to Environments that connect to the given provider on the remote side. See [`interconnect_provider`](#interconnect_provider-block) below for details.
* `location` - (Optional) Filters results to Environments that connect to the given location distinguisher. This is the location on the remote partner's side, not an AWS Region, and its format is specific to the partner (for example, a GCP region such as `us-east4`).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `interconnect_provider` Block

Exactly one of the following arguments must be set.

* `cloud_service_provider` - (Optional) Name of a cloud service provider. Connections to this provider are considered Multicloud connections.
* `last_mile_provider` - (Optional) Name of a last mile provider. Connections to this provider are considered LastMile connections.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `environments` - List of environments. See [`environments`](#environments-block) below for details.

### `environments` Block

* `activation_page_url` - URL on the Interconnect partner's portal where you confirm the connection using its activation key.
* `bandwidths` - Sets of bandwidths available and supported on this environment. See [`environments` `bandwidths`](#environments-bandwidths-block) below for details.
* `environment_id` - Identifier of the Environment.
* `interconnect_provider` - Name of the provider on the remote side of this environment.
* `location` - Provider-specific location on the remote side.
* `remote_identifier_type` - Type of identifying information that should be supplied to the `remote_account` argument of a connection for this environment. One of `account` or `email`.
* `state` - State of the Environment. One of `available`, `limited`, or `unavailable`.
* `type` - Specific product type of connections provided by this Environment.

### `environments` `bandwidths` Block

* `available` - List of currently available bandwidths.
* `supported` - List of all bandwidths that this environment plans to support.
