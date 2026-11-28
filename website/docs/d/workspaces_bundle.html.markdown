---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_bundle"
description: |-
  Retrieve information about an AWS WorkSpaces bundle.
---

# Data Source: aws_workspaces_bundle

Retrieve information about an AWS WorkSpaces bundle.

## Example Usage

### By ID

```terraform
data "aws_workspaces_bundle" "example" {
  bundle_id = "wsb-b0s22j3d7"
}
```

### By Owner & Name

```terraform
data "aws_workspaces_bundle" "example" {
  owner = "AMAZON"
  name  = "Value with Windows 10 and Office 2016"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `bundle_id` - (Optional) ID of the bundle.
* `name` - (Optional) Name of the bundle. You cannot combine this parameter with `bundle_id`.
* `owner` - (Optional) Owner of the bundles. You have to leave it blank for own bundles. You cannot combine this parameter with `bundle_id`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `compute_type` - Compute type. See supported fields below.
* `description` - Description of the bundle.
* `root_storage` - Root volume. See supported fields below.
* `user_storage` - User storage. See supported fields below.

### `compute_type` Block

* `name` - Name of the compute type.

### `root_storage` Block

* `capacity` - Size of the root volume.

### `user_storage` Block

* `capacity` - Size of the user storage.
