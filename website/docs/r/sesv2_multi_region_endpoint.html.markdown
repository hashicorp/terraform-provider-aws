---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_multi_region_endpoint"
description: |-
  Manages an AWS SESv2 (Simple Email V2) Multi Region Endpoint.
---

# Resource: aws_sesv2_multi_region_endpoint

Manages an AWS SESv2 (Simple Email V2) Multi Region Endpoint (global endpoint). Traffic is split equally between the primary region (where the resource is created) and the secondary region specified in the `details` block.

## Example Usage

### Basic Usage

```terraform
data "aws_region" "secondary" {}

resource "aws_sesv2_multi_region_endpoint" "example" {
  endpoint_name = "example"

  details {
    routes_details {
      region = data.aws_region.secondary.name
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `endpoint_name` - (Required, Forces new resource) Name of the multi-region endpoint.
* `details` - (Required, Forces new resource) Configuration details for the endpoint. See [`details`](#details) below.

### `details`

* `routes_details` - (Required) List of secondary region route configurations. Each entry contains:
  * `region` - (Required, Forces new resource) Name of the secondary AWS region.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name of the multi-region endpoint (same as `endpoint_name`).
* `endpoint_id` - ID assigned to the multi-region endpoint.
* `routes` - List of active routes. Each entry contains:
  * `region` - AWS region name for this route.
* `status` - Current status of the endpoint. One of `CREATING`, `READY`, `FAILED`, `DELETING`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 Multi Region Endpoints using the `endpoint_name`. For example:

```terraform
import {
  to = aws_sesv2_multi_region_endpoint.example
  id = "example"
}
```

Using `terraform import`, import SESv2 Multi Region Endpoints using the `endpoint_name`. For example:

```console
% terraform import aws_sesv2_multi_region_endpoint.example example
```
