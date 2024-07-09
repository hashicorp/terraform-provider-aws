---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ami_ids"
description: |-
  Provides a list of AMI IDs.
---

# Data Source: aws_service_principal

Use this data source to create a Service Principal Name for a service in a given region.

Service Principals Names should always end in the standard global format: `{servicename}.amazonaws.com` - however there 
are edge cases where this is not the case. In some AWS partitions - there are edge cases where regions expect a different 
format

## Example Usage

```terraform
data "aws_service_principal" "current" {
  service_name = "s3"
}

data "aws_service_principal" "test" {
  service_name = "s3"
  region = "us-iso-east-1"
}
```

## Argument Reference

* `service_name` - (Required) The name of the service you want to generate a Service Principal Name for.

* `region` - (Optional) Region you'd like the SPN for. By default, fetches the current region.

## Attribute Reference

`id` Identifier of the current Service Principal (compound of service, region and suffix). (e.g. `logs.us-east-1.amazonaws.com`in AWS Commercial, `logs.cn-north-1.amazonaws.com.cn` in AWS China)

`name` Service Principal Name (e.g., `logs.amazonaws.com` in AWS Commercial, `logs.amazonaws.com.cn` in AWS China)

`service` The service used for SPN generation (e.g. `logs`)

`suffix` The suffix of the SPN (e.g., `amazonaws.com` in AWS Commercial, `amazonaws.com.cn` in AWS China)

`region` - Region identifier of the generated SPN (e.g., `us-east-1` in AWS Commercial, `cn-north-1` in AWS China).
