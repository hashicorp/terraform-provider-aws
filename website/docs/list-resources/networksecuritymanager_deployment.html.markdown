---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_deployment"
description: |-
  Lists AWS Network Security Manager Deployment resources.
---

# List Resource: aws_networksecuritymanager_deployment

Lists AWS Network Security Manager Deployment resources.

## Example Usage

```terraform
list "aws_networksecuritymanager_deployment" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
