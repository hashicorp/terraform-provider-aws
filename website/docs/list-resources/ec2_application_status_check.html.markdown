---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_application_status_check"
description: |-
  Lists EC2 Application Status Check resources.
---

# List Resource: aws_ec2_application_status_check

Lists EC2 Application Status Check resources.

## Example Usage

### Basic Usage

```hcl
list "aws_ec2_application_status_check" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
