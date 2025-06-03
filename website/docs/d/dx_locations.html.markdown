---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_locations"
description: |-
  Retrieve information about the AWS Direct Connect locations in the current AWS Region.
---

# Data Source: aws_dx_locations

Retrieve information about the AWS Direct Connect locations in the current AWS Region.
These are the locations that can be specified when configuring [`aws_dx_connection`](/docs/providers/aws/r/dx_connection.html) or [`aws_dx_lag`](/docs/providers/aws/r/dx_lag.html) resources.

~> **Note:** This data source is different from the [`aws_dx_location`](/docs/providers/aws/d/dx_location.html) data source which retrieves information about a specific AWS Direct Connect location in the current AWS Region.

## Example Usage

```terraform
data "aws_dx_locations" "available" {}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `location_codes` - Code for the locations.
