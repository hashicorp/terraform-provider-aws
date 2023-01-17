---
subcategory: "Inspector V2"
layout: "aws"
page_title: "AWS: aws_inspector2_member_association"
description: |-
  Terraform resource for managing an AWS Inspector V2 Member Association.
---

# Resource: aws_inspector2_member_association

Terraform resource for associating accounts to existing Inspector2 instances.

## Example Usage

```terraform
# Main account
provider "aws" {}

# Delegated admin account
provider "aws" {
  alias  = "inspector2admin"
  assume_role {
    role_arn = "arn:aws:iam::012345678901:role/delegated-inspector-account"
  }
}

resource "aws_inspector2_delegated_admin_account" "example" {
  account_id = "012345678901"
}

# Delegated admin account needs to be enabled for dashboard to show
resource "aws_inspector2_enabler" "admin" {
  account_ids    = ["012345678901"]
  resource_types = ["EC2"]
  provider = aws.inspector2admin
  depends_on = [aws_inspector2_delegated_admin_account.example]
}

# Associating a new account to the existing inspector2 instance
resource "aws_inspector2_member_association" "new" {
  account_id    = "109876543210"
  provider = aws.inspector2admin

  # Admin account needs to be enabled
  depends_on = [aws_inspector2_enabler.admin]
}

resource "aws_inspector2_enabler" "new" {
  account_ids    = [aws_inspector2_member_association.new.account_id]
  resource_types = ["EC2"]
  provider = aws.inspector2admin
  depends_on = [aws_inspector2_member_association.new]
}
```

### Basic Usage

```terraform
resource "aws_inspector2_member_association" "example" {
  account_id = "012345678901"
}
```

## Argument Reference

The following arguments are required:

* `account_id` - (Required) ID of the account to associate

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `account_id` - (Required) ID of the account to associate