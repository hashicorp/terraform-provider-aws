---
subcategory: "AppFlow"
layout: "aws"
page_title: "AWS: aws_appflow_connector_profile"
description: |-
  Lists AppFlow Connector Profile resources.
---

# List Resource: aws_appflow_connector_profile

Lists AppFlow Connector Profile resources.

## Example Usage

```terraform
list "aws_appflow_connector_profile" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
