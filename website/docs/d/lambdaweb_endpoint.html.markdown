---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_endpoint"
description: |-
  Provides details about an AWS Lambda Web endpoint.
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Data Source: aws_lambdaweb_endpoint

Provides details about an AWS Lambda Web endpoint, including its HTTPS domain names.

## Example Usage

```terraform
data "aws_lambdaweb_endpoint" "example" {
  function_name = "example"
  endpoint_name = "default"
}
```

## Argument Reference

The following arguments are required:

* `endpoint_name` - (Required) Name of the endpoint.
* `function_name` - (Required) Name of the function the endpoint belongs to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the endpoint.
* `auth_type` - Authentication mode of the endpoint.
* `auto_deployment_mode` - Deployment mode of the endpoint.
* `domain_name` - HTTPS domain of the endpoint. Empty for `PerRegion` endpoints, which only expose `regional_domain_names`.
* `endpoint_type` - Endpoint type (`HomeRegion`, `MultiRegion` or `PerRegion`).
* `regional_domain_names` - Map of Region to that Region's independent domain name.
* `regions` - Regions the endpoint spans.
* `state` - Current state of the endpoint.
