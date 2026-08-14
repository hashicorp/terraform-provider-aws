---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_endpoint"
description: |-
  Manages an AWS Lambda Web endpoint.
---

# Resource: aws_lambdaweb_endpoint

Manages an AWS Lambda Web endpoint: an HTTPS domain for a function with its own authentication, region placement, and revision routing. A function can have up to 20 endpoints; manage the initial endpoint with the `endpoint_config` block on [`aws_lambdaweb_function`](lambdaweb_function.html.markdown) and additional endpoints with this resource.

~> **Note:** Lambda Web is available in select regions. Regions active as of July 2026: `us-east-1`, `eu-west-1`. The `us-west-2` rollout is pending.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambdaweb_endpoint" "example" {
  function_name = aws_lambdaweb_function.example.function_name
  endpoint_name = "canary"
  endpoint_type = "HomeRegion"
  auth_type     = "ApplicationManaged"
}
```

### Weighted (Canary) Routing

```terraform
resource "aws_lambdaweb_endpoint" "example" {
  function_name        = aws_lambdaweb_function.example.function_name
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
      revision_id = aws_lambdaweb_function.example.latest_revision_id
      weight      = 10
    },
  ]
}
```

## Argument Reference

The following arguments are required:

* `auth_type` - (Required) Authentication mode. Valid values are `ApplicationManaged` and `IamAuth`. (`AWS_SERVICE_AUTH` was removed in V2.) This attribute is mutable and can be updated in place.
* `endpoint_name` - (Required) Name of the endpoint, up to 64 characters. Changing this forces a new resource.
* `endpoint_type` - (Required) Endpoint type. Valid values are `HomeRegion`, `MultiRegion`, and `PerRegion`. Changing this forces a new resource.
* `function_name` - (Required) Name of the function this endpoint belongs to, up to 64 characters. Changing this forces a new resource.

The following arguments are optional:

* `auto_deployment_mode` - (Optional) Deployment mode. `LatestRevision` makes the endpoint follow the latest revision automatically; `Disabled` requires explicit `revision_weights`. `MultiRegion` endpoints require `Disabled`. Defaults to `LatestRevision`.
* `description` - (Optional) Description of the endpoint.
* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.
* `regions` - (Optional) Regions the endpoint spans (maximum 5). `MultiRegion` and `PerRegion` endpoints require at least 2 distinct regions, or none at all: the home region is added automatically. Changing this forces a new resource.
* `revision_weights` - (Optional) Traffic routing. Required when `auto_deployment_mode` is `Disabled` and must be omitted when `LatestRevision`. One or two entries; weights must sum to 100. [See below](#revision_weights-block). When a new revision is rolled on the function, update these weights to shift traffic to it — with `auto_deployment_mode = "Disabled"` traffic never moves automatically.

### `revision_weights` Block

* `revision_id` - (Required) ID of the revision to route traffic to.
* `weight` - (Required) Percentage of traffic for this revision (1-100). With one entry the weight must be 100; with two entries the weights must sum to 100.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the endpoint in the format `arn:aws:lambda:{region}:{account}:web-function/{function_name}|{endpoint_name}`.
* `domain_name` - HTTPS domain name of the endpoint. Empty for `PerRegion` endpoints, which only expose `regional_domain_names`.
* `regional_domain_names` - Map of Region to that Region's independent domain name.
* `state` - Current state of the endpoint.
* `state_reason` - Reason for the current state, useful when a regional deployment is `Pending` or `Failed`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) supports resource identity. Import a Lambda Web Endpoint using its identity attributes, `function_name` and `endpoint_name`. For example:

```terraform
import {
  to = aws_lambdaweb_endpoint.example

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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Lambda Web Endpoint using the `function_name` and `endpoint_name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_lambdaweb_endpoint.example
  id = "example,canary"
}
```

Using `terraform import`, import a Lambda Web Endpoint using the `function_name` and `endpoint_name` separated by a comma (`,`). For example:

```console
% terraform import aws_lambdaweb_endpoint.example example,canary
```
