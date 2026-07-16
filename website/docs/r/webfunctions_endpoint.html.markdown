---
subcategory: "Lambda Web Functions"
layout: "aws"
page_title: "AWS: aws_webfunctions_endpoint"
description: |-
  Manages an AWS Lambda Web Functions endpoint.
---

# Resource: aws_webfunctions_endpoint

Manages an AWS Lambda Web Functions endpoint: an HTTPS domain for a function with its own authentication, region placement, and revision routing. A function can have up to 20 endpoints; manage the initial endpoint with the `endpoint_config` block on [`aws_webfunctions_function`](webfunctions_function.html.markdown) and additional endpoints with this resource.

~> **Note:** Lambda Web Functions is available in select regions. Regions active as of July 2026: `us-east-1`, `eu-west-1`. The `us-west-2` rollout is pending.

## Example Usage

### Basic Usage

```terraform
resource "aws_webfunctions_endpoint" "example" {
  function_name = aws_webfunctions_function.example.function_name
  endpoint_name = "canary"
  endpoint_type = "HomeRegion"
  auth_type     = "ApplicationManaged"
}
```

### Weighted (Canary) Routing

```terraform
resource "aws_webfunctions_endpoint" "example" {
  function_name        = aws_webfunctions_function.example.function_name
  endpoint_name        = "prod"
  endpoint_type        = "MultiRegion"
  regions              = ["us-east-1", "eu-west-1"]
  auth_type            = "IamAuth"
  auto_deployment_mode = "Disabled"

  revision_weights = [
    {
      revision_id = "rev-aaaaaaaa"
      weight      = 90
    },
    {
      revision_id = aws_webfunctions_function.example.latest_revision_id
      weight      = 10
    },
  ]
}
```

## Argument Reference

The following arguments are required:

* `auth_type` - (Required) Authentication mode. Valid values are `ApplicationManaged` and `IamAuth`. (`AWS_SERVICE_AUTH` was removed in V2.) This attribute is mutable and can be updated in place.
* `endpoint_name` - (Required) Name of the endpoint. Changing this forces a new resource.
* `endpoint_type` - (Required) Endpoint type. Valid values are `HomeRegion`, `MultiRegion`, and `PerRegion`. Changing this forces a new resource.
* `function_name` - (Required) Name of the function this endpoint belongs to. Changing this forces a new resource.

The following arguments are optional:

* `auto_deployment_mode` - (Optional) Deployment mode. `LatestRevision` makes the endpoint follow the latest revision automatically; `Disabled` requires explicit `revision_weights`. `MultiRegion` endpoints require `Disabled`. Defaults to `LatestRevision`.
* `description` - (Optional) Description of the endpoint.
* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.
* `regions` - (Optional) Regions the endpoint spans (maximum 17). Changing this forces a new resource.
* `revision_weights` - (Optional) Traffic routing. Required when `auto_deployment_mode` is `Disabled` and must be omitted when `LatestRevision`. One or two entries; weights must sum to 100. [See below](#revision_weights-block).

### `revision_weights` Block

* `revision_id` - (Required) ID of the revision to route traffic to.
* `weight` - (Required) Percentage of traffic for this revision (1-100). With one entry the weight must be 100; with two entries the weights must sum to 100.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the endpoint in the format `arn:aws:lambda:{region}:{account}:web-function/{function_name}|{endpoint_name}`.
* `domain_name` - HTTPS domain name of the endpoint.
* `state` - Current state of the endpoint.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) supports resource identity. Import a Web Functions Endpoint using its identity attributes, `function_name` and `endpoint_name`. For example:

```terraform
import {
  to = aws_webfunctions_endpoint.example

  identity = {
    function_name = "example"
    endpoint_name = "canary"
  }
}
```

### Identity Schema

#### Required

* `endpoint_name` (String) Name of the endpoint.
* `function_name` (String) Name of the function.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Web Functions Endpoint using the `function_name` and `endpoint_name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_webfunctions_endpoint.example
  id = "example,canary"
}
```

Using `terraform import`, import a Web Functions Endpoint using the `function_name` and `endpoint_name` separated by a comma (`,`). For example:

```console
% terraform import aws_webfunctions_endpoint.example example,canary
```
