---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_portfolio_share"
description: |-
  Manages a Service Catalog Portfolio Share
---

# Resource: aws_servicecatalog_portfolio_share

Manages a Service Catalog Portfolio Share. Shares the specified portfolio with the specified account or organization node. You can share portfolios to an organization, an organizational unit, or a specific account.

If the portfolio share with the specified account or organization node already exists, using this resource to re-create the share will have no effect and will not return an error. You can then use this resource to update the share.

~> **NOTE:** Shares to an organization node can only be created by the management account of an organization or by a delegated administrator. If a delegated admin is de-registered, they can no longer create portfolio shares.

~> **NOTE:** AWSOrganizationsAccess must be enabled in order to create a portfolio share to an organization node.

~> **NOTE:** You can't share a shared resource, including portfolios that contain a shared product.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_portfolio_share" "example" {
  account_id   = "012128675309"
  portfolio_id = aws_servicecatalog_portfolio.example.id
}
```

## Argument Reference

The following arguments are required:

* `account_id` - (Required if `organization_node` is not included) AWS account ID. For example, `123456789012`.
* `organization_node` - (Required if `account_id` is not included) Configuration block with an organization node to whom you are going to share. Detailed below.
* `portfolio_id` - (Required) Portfolio identifier.
* `type` - (Required) Type of portfolio share. Valid values are `ACCOUNT` (an external account), `ORGANIZATION` (a share to every account in an organization), `ORGANIZATIONAL_UNIT`, `ORGANIZATION_MEMBER_ACCOUNT` (a share to an account in an organization).

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `share_tag_options` - (Optional) Whether to enable sharing of `aws_servicecatalog_tag_option` resources when creating the portfolio share.

### organization_node

The following arguments are supported:

* `type` - (Optional) Organization node type. Valid values are `ACCOUNT` (this value corresponds to `ORGANIZATION_MEMBER_ACCOUNT`), `ORGANIZATION`, and `ORGANIZATIONAL_UNIT`.
* `value` - (Optional) Identifier of the organization node.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `accepted` - Whether the shared portfolio is imported by the recipient account. If the recipient is in an organization node, the share is automatically imported, and the field is always set to true.

## Import

`aws_servicecatalog_portfolio_share` can be imported using the portfolio share ID, e.g.

```
$ terraform import aws_servicecatalog_portfolio_share.example arn:aws:catalog:us-east-1:123456789012:portfolio share/prod-dnigbtea24ste
```
