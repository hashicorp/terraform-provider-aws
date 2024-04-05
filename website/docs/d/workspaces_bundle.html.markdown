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

* `bundle_id` – (Optional) ID of the bundle.
* `owner` – (Optional) Owner of the bundles. You have to leave it blank for own bundles. You cannot combine this parameter with `bundle_id`.
* `name` – (Optional) Name of the bundle. You cannot combine this parameter with `bundle_id`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` – The description of the bundle.
* `bundle_id` – The ID of the bundle.
* `name` – The name of the bundle.
* `owner` – The owner of the bundle.
* `compute_type` – The compute type. See supported fields below.
* `root_storage` – The root volume. See supported fields below.
* `user_storage` – The user storage. See supported fields below.

### `compute_type`

* `name` - Name of the compute type.

### `root_storage`

* `capacity` - Size of the root volume.

### `user_storage`

* `capacity` - Size of the user storage.
