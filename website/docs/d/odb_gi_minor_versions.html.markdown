---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_gi_minor_versions"
description: |-
  Provides details about available Oracle Database@AWS GI minor versions.
---

# Data Source: aws_odb_gi_minor_versions

Provides details about available Oracle Database@AWS GI minor versions and their Grid Infrastructure software image IDs.

## Example Usage

### Basic Usage

```terraform
data "aws_odb_gi_minor_versions" "example" {
  availability_zone_id = "use1-az6"
  gi_version            = "19.0.0.0"
  shape_family          = "EXADB_XS"
}
```

## Argument Reference

The following arguments are required:

* `gi_version` - (Required) GI major version. Length must be between `1` and `255` characters.
* `shape_family` - (Required) Shape family for the GI minor versions. Length must be between `1` and `255` characters.

The following arguments are optional:

* `availability_zone` - (Optional) Availability Zone to filter GI minor versions and retrieve Grid Infrastructure software image IDs. Length must be between `1` and `255` characters.
* `availability_zone_id` - (Optional) Availability Zone ID to filter GI minor versions and retrieve Grid Infrastructure software image IDs. Length must be between `1` and `255` characters.
* `region` - (Optional) Region where this data source will be [read](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `gi_minor_versions` - Available GI minor versions and their Grid Infrastructure software image IDs. See [`gi_minor_versions`](#gi_minor_versions) below.

### `gi_minor_versions` Block

* `grid_image_id` - Grid Infrastructure software image ID. The value can be null when neither `availability_zone` nor `availability_zone_id` is specified.
* `version` - GI minor version.
