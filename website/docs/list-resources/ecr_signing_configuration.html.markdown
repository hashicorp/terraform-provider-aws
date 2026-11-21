---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_signing_configuration"
description: |-
  Lists the ECR (Elastic Container Registry) signing configuration.
---

# List Resource: aws_ecr_signing_configuration

Lists the ECR (Elastic Container Registry) signing configuration.

A registry has at most one signing configuration per Region, so a query returns at
most one result per queried Region.

## Example Usage

```terraform
list "aws_ecr_signing_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
