---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_api_key"
description: |-
  Terraform resource for managing an AWS Location Service API Key.
---

# Resource: aws_location_api_key

Terraform resource for managing an AWS Location Service API Key.

~> **Note:** Deleting this resource uses `ForceDelete`, which bypasses the normal 90-day deactivation period required by the API. Be cautious when destroying production keys that may still be in use.

## Example Usage

### Basic Usage

```terraform
resource "aws_location_api_key" "example" {
  key_name  = "example"
  no_expiry = true

  restrictions {
    allow_actions   = ["geo:GetMap*"]
    allow_resources = ["arn:aws:geo:*:*:map/*"]
  }
}
```

### With Expiry Time

```terraform
resource "aws_location_api_key" "example" {
  key_name    = "example"
  description = "My API key"
  expire_time = "2025-12-31T23:59:59Z"

  restrictions {
    allow_actions   = ["geo:GetMap*", "geo:SearchPlaceIndexForText"]
    allow_resources = ["arn:aws:geo:*:*:map/*", "arn:aws:geo:*:*:place-index/*"]
  }

  tags = {
    Environment = "production"
  }
}
```

### With Allowed Referers

```terraform
resource "aws_location_api_key" "example" {
  key_name  = "example"
  no_expiry = true

  restrictions {
    allow_actions   = ["geo:GetMap*"]
    allow_resources = ["arn:aws:geo:*:*:map/*"]
    allow_referers  = ["https://example.com/*"]
  }
}
```

## Argument Reference

The following arguments are required:

* `key_name` - (Required, Forces new resource) A custom name for the API key. Must contain only alphanumeric characters, hyphens, periods, and underscores. Must be unique within the account.
* `restrictions` - (Required) API key restrictions. See [Restrictions](#restrictions) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) An optional description for the API key.
* `expire_time` - (Optional) The timestamp for when the API key expires, in RFC3339 format (e.g., `2025-12-31T23:59:59Z`). Conflicts with `no_expiry`. If neither `expire_time` nor `no_expiry` is set, the key defaults to no expiry.
* `no_expiry` - (Optional) Set to `true` to create an API key with no expiration. Conflicts with `expire_time`. Defaults to `true` when neither `expire_time` nor `no_expiry` is specified.
* `tags` - (Optional) Key-value tags for the API key. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Restrictions

* `allow_actions` - (Required) A list of allowed actions. Example: `["geo:GetMap*"]`. See [Amazon Location Service API Actions](https://docs.aws.amazon.com/location/latest/developerguide/using-apikeys.html) for valid values.
* `allow_resources` - (Required) A list of allowed resource ARNs. Supports wildcards. Example: `["arn:aws:geo:*:*:map/*"]`.
* `allow_referers` - (Optional) A list of allowed HTTP referer patterns. Supports wildcards.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `create_time` - The timestamp for when the API key was created in ISO 8601 format.
* `key_arn` - The Amazon Resource Name (ARN) for the API key resource.
* `key_value` - The actual API key value. This value is **sensitive** and will not be shown in plan output.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_time` - The timestamp for when the API key was last updated in ISO 8601 format.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Location Service API Keys using the `key_name`. For example:

```terraform
import {
  to = aws_location_api_key.example
  id = "example"
}
```

Using `terraform import`, import Location Service API Keys using the `key_name`. For example:

```console
% terraform import aws_location_api_key.example example
```
