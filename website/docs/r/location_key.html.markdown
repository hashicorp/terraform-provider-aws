---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_key"
description: |-
  Terraform resource for managing an AWS Location Key.
---
# Resource: aws_location_key

Terraform resource for managing an AWS Location Key.

## Example Usage

### Basic Usage

```terraform
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = "example"
}

resource "aws_location_key" "example" {
  key_name = "example_key_name"
  no_expiry = true

  restrictions {
    allow_actions = [
      "geo:SearchPlaceIndexForText",
      "geo:SearchPlaceIndexForPosition",
      "geo:SearchPlaceIndexForSuggestions",
      "geo:GetPlace",
    ]
    allow_resources = [
      aws_location_place_index.test.index_arn,
    ]
  }
}
```

## Argument Reference

The following arguments are required:

* `key_name` - (Required) Name of the API key.
* `restrictions` - (Required) Restrictions for the API key. See [`restrictions`](#restrictions-argument-reference) below.

The following arguments are optional:

* `description` - (Optional) Description for the API key.
* `expire_time` - (Optional) Timestamp for when the API key will expire in ISO 8601 format. Required if `no_expiry` is not set to true.
* `force_update` - (Optional) Specify true if you are updating the ExpireTime or Restrictions arguments for an API key that has been used in the past 7 days.
* `no_expiry` - (Optional) Specify true to set no expiration time for the API key. Required if `expire_time` is not set.
* `tags` - (Optional) Key-value tags for the API key. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `restrictions` Argument Reference

The following arguments are required:

* `allow_actions` - (Required) List of allowed actions that an API key resource grants permissions to perform. Must have at least one action for each type of resource. See [API reference](https://docs.aws.amazon.com/location/latest/APIReference/API_ApiKeyRestrictions.html) for more details.
* `allow_resources` - (Required) List of allowed resource ARNs that a API key bearer can perform actions on.

The following arguments are optional:

* `allow_referers` - (Optional) List of allowed HTTP referers for which requests must originate from. Requests using this API key from other domains will not be allowed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `create_time` - Timestamp for when the API key was created in ISO 8601 format.
* `key` - String (key value) of an API key. Used when making API calls to authorize the call.
* `key_arn` - Amazon Resource Name (ARN) of the API key.
* `tags_all` - Map of tags assigned to the API key, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_time` - Timestamp for when the API key was last updated in ISO 8601 format.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Location Key using the `example_id_arg`. For example:

```terraform
import {
  to = aws_location_key.example
  id = "key-id-12345678"
}
```

Using `terraform import`, import Location Key using the `example_id_arg`. For example:

```console
% terraform import aws_location_key.example key-id-12345678
```
