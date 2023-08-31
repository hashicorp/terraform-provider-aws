---
subcategory: "License Manager"
layout: "aws"
page_title: "AWS: aws_licensemanager_grants"
description: |-
    Get information about a set of license manager grant licenses
---

# Data Source: aws_licensemanager_grants

This resource can be used to get a set of license grant ARNs matching a filter.

## Example Usage

The following shows getting all license grant ARNs granted to your account.

```terraform
data "aws_caller_identity" "current" {}

data "aws_licensemanager_grants" "test" {
  filter {
    name = "GranteePrincipalARN"
    values = [
      "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
    ]
  }
}
```

## Argument Reference

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/license-manager/latest/APIReference/API_ListReceivedGrants.html#API_ListReceivedGrants_RequestSyntax).
  For example, if filtering using `ProductSKU`, use:

```terraform
data "aws_licensemanager_grants" "selected" {
  filter {
    name   = "ProductSKU"
    values = [""] # insert values here
  }
}
```

* `values` - (Required) Set of values that are accepted for the given field.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - List of all the license grant ARNs found.
