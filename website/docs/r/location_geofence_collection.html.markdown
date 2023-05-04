---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_geofence_collection"
description: |-
  Terraform resource for managing an AWS Location Geofence Collection.
---

# Resource: aws_location_geofence_collection

Terraform resource for managing an AWS Location Geofence Collection.

## Example Usage

```terraform
resource "aws_location_geofence_collection" "example" {
  collection_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `collection_name` - (Required) The name of the geofence collection.

The following arguments are optional:

* `description` - (Optional) The optional description for the geofence collection.
* `kms_key_id` - (Optional) A key identifier for an AWS KMS customer managed key assigned to the Amazon Location resource.
* `tags` - (Optional) Key-value tags for the geofence collection. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `collection_arn` - The Amazon Resource Name (ARN) for the geofence collection resource. Used when you need to specify a resource across all AWS.
* `create_time` - The timestamp for when the geofence collection resource was created in ISO 8601 format.
* `update_time` - The timestamp for when the geofence collection resource was last updated in ISO 8601 format.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

Location Geofence Collection can be imported using the `collection_name`, e.g.,

```
$ terraform import aws_location_geofence_collection.example example
```
