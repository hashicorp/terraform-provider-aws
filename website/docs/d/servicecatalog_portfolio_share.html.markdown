---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_portfolio"
description: |-
  Provides information for a Service Catalog Portfolio Share.
---

# Data Source: aws_servicecatalog_portfolio

Provides information for a Service Catalog Portfolio.

## Example Usage

```terraform
data "aws_servicecatalog_portfolio_share" "example" {
  portfolio_id = "port-00000000"
  type         = "ACCOUNT"
}
```

## Argument Reference

The following arguments are required:

* `portfolio_id` - (Required) Portfolio identifier.
* `type` - (Required) Type of Portfolio.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* "accepted" - Whether each principal identified is an accepted share or not.
* "id" - Id of the portfolio.
* "portfolio_id" - Same as id above, duplicate is required.
* "principal_ids" - Array of principal ids portfolio has been shared with.
* "share_principals" - Boolean array indicating which have share_principals enabled.
* "type" - Array of strings indicating type of each share.
