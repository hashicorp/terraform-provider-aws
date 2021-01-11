---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_service"
description: |-
  Construct AWS service names
---

# Data Source: aws_service

Use this data source to compose AWS service names.

## Example Usage

```hcl
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

data "aws_service" "s3" {
  service = "s3"
}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = data.aws_service.s3.id
}
```

## Argument Reference

~> **Note:** If the `name` argument is provided, `region`, `service`, and `service_prefix` values are ignored.

* `name` - Name of the service (e.g. `com.amazonaws.us-west-2.s3`)
* `region` - Region of the service (e.g. `us-west-2`, `ap-northeast-1`).
* `service` - Service (e.g. `s3`, `ec2`). Defaults to `ec2`.
* `service_prefix` - Prefix of the service (e.g. `com.amazonaws` in AWS Commercial, `cn.com.amazonaws` in AWS China).

## Attributes Reference

Besides the arguments above, no attributes are exported.
