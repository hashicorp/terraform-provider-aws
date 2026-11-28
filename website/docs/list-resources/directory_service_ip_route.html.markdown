---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_ip_route"
description: |-
  Lists Directory Service IP Route resources.
---

# List Resource: aws_directory_service_ip_route

Lists Directory Service IP Route resources for a directory.

## Example Usage

```terraform
list "aws_directory_service_ip_route" "example" {
  provider = aws

  config {
    directory_id = "d-1234567890"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `directory_id` - (Required) Identifier of the directory whose IP routes to list.
* `region` - (Optional) Region to query. Defaults to provider region.
