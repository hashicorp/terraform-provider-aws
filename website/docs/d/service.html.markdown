---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_service"
description: |-
  Compose and decompose AWS service DNS names
---

# Data Source: aws_service

Use this data source to compose and decompose AWS service DNS names.

## Example Usage

### Get Service DNS Name

```hcl
data "aws_region" "current" {}

data "aws_service" "test" {
  region     = data.aws_region.current.name
  service_id = "ec2"
}
```

### Use Service Reverse DNS Name to Get Components

```hcl
data "aws_service" "s3" {
  reverse_dns_name = "cn.com.amazonaws.cn-north-1.s3"
}
```

### Determine Regional Support for a Service

```hcl
data "aws_service" "s3" {
  reverse_dns_name = "com.amazonaws.us-gov-west-1.waf"
}
```

## Argument Reference

The following arguments are optional:

* `dns_name` - (Optional) DNS name of the service (_e.g.,_ `rds.us-east-1.amazonaws.com`). One of `dns_name`, `reverse_dns_name`, or `service_id` is required.
* `partition` - (Optional) Partition corresponding to the region.
* `region` - (Optional) Region of the service (_e.g.,_ `us-west-2`, `ap-northeast-1`).
* `reverse_dns_name` - (Optional) Reverse DNS name of the service (_e.g.,_ `com.amazonaws.us-west-2.s3`). One of `dns_name`, `reverse_dns_name`, or `service_id` is required.
* `reverse_dns_prefix` - (Optional) Prefix of the service (_e.g.,_ `com.amazonaws` in AWS Commercial, `cn.com.amazonaws` in AWS China).
* `service_id` - (Optional) Service (_e.g.,_ `s3`, `rds`, `ec2`). One of `dns_name`, `reverse_dns_name`, or `service_id` is required.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `supported` - Whether the service is supported in the region's partition. New services may not be listed immediately as supported.
