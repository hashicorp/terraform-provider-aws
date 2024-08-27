---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_provisioning_artifacts"
description: |-
  Provides information on Service Catalog Provisioning Artifacts
---

# Data Source: aws_servicecatalog_provisioning_artifacts

Lists the provisioning artifacts for the specified product.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalog_provisioning_artifacts" "example" {
  product_id = "prod-yakog5pdriver"
}
```

## Argument Reference

The following arguments are required:

* `product_id` - (Required) Product identifier.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `provisioning_artifact_details` - List with information about the provisioning artifacts. See details below.

### provisioning_artifact_details

* `active` - Indicates whether the product version is active.
* `created_time` - The UTC time stamp of the creation time.
* `description` - The description of the provisioning artifact.
* `guidance` - Information set by the administrator to provide guidance to end users about which provisioning artifacts to use.
* `id` - The identifier of the provisioning artifact.
* `name` - The name of the provisioning artifact.
* `type` - The type of provisioning artifact.
