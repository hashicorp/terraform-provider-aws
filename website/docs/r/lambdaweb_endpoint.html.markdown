---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_endpoint"
description: |-
  Manages an AWS Lambda Web endpoint.
---

# Resource: aws_lambdaweb_endpoint

Manages an AWS Lambda Web endpoint: an HTTPS domain for a function with its own authentication, region placement, and revision routing. A function can have up to 10 endpoints; manage the initial endpoint with the `endpoint_config` block on [`aws_lambdaweb_function`](lambdaweb_function.html.markdown) and additional endpoints with this resource.

~> **Note:** Lambda Web is available in select regions. As of August 2026 the API is active in 17 commercial regions, including `us-east-1` and `eu-west-1`. In regions where the service is not yet deployed, API calls fail with `AccessDeniedException`.

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

  revision_weights {
    revision_id = "rev-aaaaaaaa"
    weight      = 90
  }

  revision_weights {
    revision_id = aws_lambdaweb_function.example.latest_revision_id
    weight      = 10
  }
}
```

### Behind CloudFront

An endpoint is a plain HTTPS origin, so a distribution can front it to add edge
caching for static paths and terminate TLS closer to the viewer. Use
`domain_name` as the origin: on a `MultiRegion` endpoint it routes each request
to the nearest deployed Region, and `regional_domain_names` is empty because the
service does not publish per-Region domains for that endpoint type.

```terraform
resource "aws_lambdaweb_endpoint" "app" {
  function_name        = aws_lambdaweb_function.example.function_name
  endpoint_name        = "public"
  endpoint_type        = "MultiRegion"
  regions              = ["us-east-1", "eu-west-1"]
  auth_type            = "ApplicationManaged"
  auto_deployment_mode = "Disabled"

  revision_weights {
    revision_id = aws_lambdaweb_function.example.latest_revision_id
    weight      = 100
  }
}

data "aws_cloudfront_cache_policy" "disabled" {
  name = "Managed-CachingDisabled"
}

data "aws_cloudfront_origin_request_policy" "all_viewer_except_host" {
  name = "Managed-AllViewerExceptHostHeader"
}

resource "aws_cloudfront_distribution" "app" {
  enabled = true

  origin {
    origin_id   = "lambdaweb"
    domain_name = aws_lambdaweb_endpoint.app.domain_name

    custom_origin_config {
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
      http_port              = 80
      https_port             = 443
    }
  }

  default_cache_behavior {
    target_origin_id       = "lambdaweb"
    viewer_protocol_policy = "redirect-to-https"

    allowed_methods = ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"]
    cached_methods  = ["GET", "HEAD"]

    # The endpoint routes on its own domain, so the viewer Host header must not
    # be forwarded, and dynamic responses must not be cached.
    cache_policy_id          = data.aws_cloudfront_cache_policy.disabled.id
    origin_request_policy_id = data.aws_cloudfront_origin_request_policy.all_viewer_except_host.id
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
```

~> **Note:** Do not set `default_root_object` on a distribution that fronts a
web function. CloudFront rewrites `/` to that object after viewer-request
functions run, which frameworks that own their routing answer with a 404.

Streaming survives the extra hop: a Server-Sent Events response is relayed
without buffering. Origin groups, on the other hand, reject behaviors that allow
write methods, so a failover group can only serve the read paths and writes need
a behavior pointing at a single origin. A `PerRegion` endpoint is the one that
publishes `regional_domain_names`, which is what an origin group needs for its
primary and failover members.

## Argument Reference

The following arguments are required:

* `auth_type` - (Required) Authentication mode. Valid values are `ApplicationManaged` and `IamAuth`. The authentication type is chosen at creation and cannot be edited afterward, so changing it forces a new resource.
* `endpoint_name` - (Required) Name of the endpoint, up to 64 characters. Changing this forces a new resource.
* `endpoint_type` - (Required) Endpoint type. Valid values are `HomeRegion`, `MultiRegion`, and `PerRegion`. Changing this forces a new resource.
* `function_name` - (Required) Name of the function this endpoint belongs to, up to 64 characters. Changing this forces a new resource.

The following arguments are optional:

* `auto_deployment_mode` - (Optional) Deployment mode. `LatestRevision` makes the endpoint follow the latest revision automatically; `Disabled` requires explicit `revision_weights`. `MultiRegion` and `PerRegion` endpoints require `Disabled`. Defaults to `LatestRevision`.
* `description` - (Optional) Description of the endpoint.
* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.
* `regions` - (Optional) Regions the endpoint spans (maximum 5). `MultiRegion` and `PerRegion` endpoints require at least 2 distinct regions, or none at all: the home region is added automatically. Changing this forces a new resource.
* `revision_weights` - (Optional) Traffic routing. Required when `auto_deployment_mode` is `Disabled` and must be omitted when `LatestRevision`. One or two entries; weights must sum to 100. [See below](#revision_weights-block). When a new revision is rolled on the function, update these weights to shift traffic to it — with `auto_deployment_mode = "Disabled"` traffic never moves automatically.
* `scaling_config` - (Optional) Scaling limits for the endpoint. When unset, the service applies account-level defaults and reports no value. [See below](#scaling_config-attribute).
* `throttle_config` - (Optional) Request throttling for the endpoint. When unset, the service applies account-level defaults and reports no value. [See below](#throttle_config-attribute).

### `revision_weights` Block

* `revision_id` - (Required) ID of the revision to route traffic to.
* `weight` - (Required) Percentage of traffic for this revision (1-100). With one entry the weight must be 100; with two entries the weights must sum to 100.

### `scaling_config` Attribute

An object (assigned with `=`, not a block):

* `max_environments` - (Required) Maximum number of concurrent execution environments the endpoint may scale to, minimum 2.

### `throttle_config` Attribute

An object (assigned with `=`, not a block):

* `rate_limit` - (Required) Maximum request rate for the endpoint, in requests per second. The service only accepts quantized values: `0`, `100`-`1000` in steps of 100, and `2000`-`10000` in steps of 1000.

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
