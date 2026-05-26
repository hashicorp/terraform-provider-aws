---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_network_insights_access_scope"
description: |-
  Lists EC2 (Elastic Compute Cloud) Network Insights Access Scope resources.
---

# List Resource: aws_ec2_network_insights_access_scope

Lists EC2 (Elastic Compute Cloud) Network Insights Access Scope resources.

## Example Usage

```terraform
list "aws_ec2_network_insights_access_scope" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
