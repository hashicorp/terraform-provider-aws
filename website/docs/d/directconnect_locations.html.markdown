---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_directconnect_locations"
description: |-
  Retrieve information about the AWS Direct Connect locations in the current AWS Region.
---

# Data Source: aws_directconnect_locations

Retrieve information about the AWS Direct Connect locations in the current AWS Region.
These are the locations that can be specified when configuring [`aws_directconnect_connection`](/docs/providers/aws/r/dx_connection.html) or [`aws_directconnect_lag`](/docs/providers/aws/r/dx_lag.html) resources.

~> **Note:** This data source is different from the [`aws_directconnect_location`](/docs/providers/aws/d/dx_location.html) data source which retrieves information about a specific AWS Direct Connect location in the current AWS Region.

## Example Usage

```hcl
data "aws_directconnect_locations" "available" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `location_codes` - The code for the locations.
