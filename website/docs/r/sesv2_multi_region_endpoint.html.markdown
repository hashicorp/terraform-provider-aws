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
resource "aws_sesv2_multi_region_endpoint" "example" {
  endpoint_name = "example"

  details {
    routes_details {
      region = "example-alternate-region"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `details` - (Required) Configuration details for the endpoint. See [`details` Block](#details-block) below.
* `endpoint_name` - (Required) Name of the multi-region endpoint.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `details` Block

The `details` block supports:

* `routes_details` - (Required) Secondary region route configuration. See [`routes_details` Block](#routes_details-block) below.

### `routes_details` Block

The `routes_details` block supports:

* `region` - (Required) Name of the secondary AWS region.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the multi-region endpoint.
* `endpoint_id` - ID assigned to the multi-region endpoint.
* `routes` - List of active routes. See [`routes` Block](#routes-block) below.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `routes` Block

The `routes` block supports:

* `region` - AWS region name for this route.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)
