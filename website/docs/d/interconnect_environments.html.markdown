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
  location = "example-location"
}
```

## Argument Reference

The following arguments are optional:

* `location` - (Optional) Filters results to Environments that connect to the given location distinguisher.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `environments` - List of environments. [See below](#environments).

### environments

* `activation_page_url` - URL on the Interconnect partner's portal where you confirm the connection using its activation key.
* `bandwidths` - Sets of bandwidths available and supported on this environment.
    * `available` - List of currently available bandwidths.
    * `supported` - List of all bandwidths that this environment plans to support.
* `environment_id` - Identifier of the Environment.
* `interconnect_provider` - Name of the provider on the remote side of this environment.
* `location` - Provider-specific location on the remote side.
* `remote_identifier_type` - Type of identifying information that should be supplied to the `remote_account` argument of a connection for this environment.
* `state` - State of the Environment.
* `type` - Specific product type of connections provided by this Environment.
