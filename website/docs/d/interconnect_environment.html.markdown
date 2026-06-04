---
subcategory: "Interconnect"
layout: "aws"
page_title: "AWS: aws_interconnect_environment"
description: |-
  Terraform data source for an AWS Interconnect Environment.
---

# Data Source: aws_interconnect_environment

Terraform data source for an AWS Interconnect Environment. An Environment defines the partner and the remote location or region to which an AWS Interconnect connection can be made.

## Example Usage

### Basic Usage

```terraform
data "aws_interconnect_environment" "example" {
  environment_id = "example-environment-id"
}
```

## Argument Reference

The following arguments are required:

* `environment_id` - (Required) Identifier of the Environment.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `activation_page_url` - URL on the Interconnect partner's portal where you confirm the connection using its activation key.
* `bandwidths` - Sets of bandwidths available and supported on this environment. [See below](#bandwidths).
* `interconnect_provider` - Name of the provider on the remote side of this environment.
* `location` - Provider-specific location on the remote side.
* `remote_identifier_type` - Type of identifying information that should be supplied to the `remote_account` argument of a connection for this environment.
* `state` - State of the Environment. One of `available`, `limited`, or `unavailable`.
* `type` - Specific product type of connections provided by this Environment.

### bandwidths

* `available` - List of currently available bandwidths.
* `supported` - List of all bandwidths that this environment plans to support.
