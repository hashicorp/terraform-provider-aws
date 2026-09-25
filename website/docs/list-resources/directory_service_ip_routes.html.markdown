---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_ip_routes"
description: |-
  Lists Directory Service IP routes resources.
---

# List Resource: aws_directory_service_ip_routes

Lists Directory Service IP routes resources. One result is returned for each directory that has IP routes configured.

## Example Usage

```terraform
list "aws_directory_service_ip_routes" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
