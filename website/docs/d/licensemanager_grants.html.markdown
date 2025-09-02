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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Custom filter block as described below.

### `filter`

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

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
