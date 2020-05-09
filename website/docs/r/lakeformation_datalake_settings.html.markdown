---
subcategory: "LakeFormation"
layout: "aws"
page_title: "AWS: aws_lakeformation_datalake_settings"
description: |-
  Manages the data lake settings for the current account
---

# Resource: aws_lakeformation_datalake_settings

Manages the data lake settings for the current account.

## Example Usage

```hcl
data "aws_iam_user" "existing_user" {
  user_name = "an_existing_user_name"
}

data "aws_iam_role" "existing_role" {
  name = "an_existing_role_name"
}

resource "aws_lakeformation_datalake_settings" "example" {
  admins = [
    "${aws_iam_user.existing_user.arn}",
    "${aws_iam_user.existing_role.arn}",
  ]
}
```

## Argument Reference

The following arguments are required:

* `admins` – (Required) A list of up to 10 AWS Lake Formation principals (users or roles).

The following arguments are optional:

* `catalog_id` – (Optional) The identifier for the Data Catalog. By default, the account ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Resource identifier with the pattern `lakeformation:settings:ACCOUNT_ID`.
