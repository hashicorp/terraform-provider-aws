---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_application_status_check_association"
description: |-
  Lists EC2 Application Status Check Association resources.
---

# List Resource: aws_ec2_application_status_check_association

Lists EC2 Application Status Check Association resources.

## Example Usage

### Basic Usage

```hcl
list "aws_ec2_application_status_check_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
