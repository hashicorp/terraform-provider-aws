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

```terraform
data "aws_region" "current" {}

data "aws_service" "test" {
  region     = data.aws_region.current.name
  service_id = "ec2"
}
```

### Use Service Reverse DNS Name to Get Components

```terraform
data "aws_service" "s3" {
  reverse_dns_name = "cn.com.amazonaws.cn-north-1.s3"
}
```

### Determine Regional Support for a Service

```terraform
data "aws_service" "s3" {
  reverse_dns_name = "com.amazonaws.us-gov-west-1.waf"
}
```

## Argument Reference

The following arguments are optional:

* `dns_name` - (Optional) DNS name of the service (_e.g.,_ `rds.us-east-1.amazonaws.com`). One of `dns_name`, `reverse_dns_name`, or `service_id` is required.
* `partition` - (Optional) Partition corresponding to the Region.
* `region` - (Optional) Region of the service (_e.g.,_ `us-west-2`, `ap-northeast-1`). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `reverse_dns_name` - (Optional) Reverse DNS name of the service (_e.g.,_ `com.amazonaws.us-west-2.s3`). One of `dns_name`, `reverse_dns_name`, or `service_id` is required.
* `reverse_dns_prefix` - (Optional) Prefix of the service (_e.g.,_ `com.amazonaws` in AWS Commercial, `cn.com.amazonaws` in AWS China).
* `service_id` - (Optional) Service endpoint ID (_e.g.,_ `s3`, `rds`, `ec2`). One of `dns_name`, `reverse_dns_name`, or `service_id` is required. A service's endpoint ID can be found in the [_AWS General Reference_](https://docs.aws.amazon.com/general/latest/gr/aws-service-information.html).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `supported` - Whether the service is supported in the region's partition. New services may not be listed immediately as supported.
