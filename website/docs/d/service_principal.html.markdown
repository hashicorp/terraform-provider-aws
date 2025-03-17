---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_service_principal"
description: |-
  Compose a Service Principal Name.
---

# Data Source: aws_service_principal

Use this data source to create a Service Principal Name for a service in a given region. Service Principal Names should always end in the standard global format: `{servicename}.amazonaws.com`. However, in some AWS partitions, AWS may expect a different format.

## Example Usage

```terraform
data "aws_service_principal" "current" {
  service_name = "s3"
}

data "aws_service_principal" "test" {
  service_name = "s3"
  region       = "us-iso-east-1"
}
```

## Argument Reference

* `service_name` - (Required) Name of the service you want to generate a Service Principal Name for.
* `region` - (Optional) Region you'd like the SPN for. By default, uses the current region.

## Attribute Reference

* `id` - Identifier of the current Service Principal (compound of service, region and suffix). (e.g. `logs.us-east-1.amazonaws.com`in AWS Commercial, `logs.cn-north-1.amazonaws.com.cn` in AWS China).
* `name` - Service Principal Name (e.g., `logs.amazonaws.com` in AWS Commercial, `logs.amazonaws.com.cn` in AWS China).
* `service` - Service used for SPN generation (e.g. `logs`).
* `suffix` - Suffix of the SPN (e.g., `amazonaws.com` in AWS Commercial, `amazonaws.com.cn` in AWS China).
*`region` - Region identifier of the generated SPN (e.g., `us-east-1` in AWS Commercial, `cn-north-1` in AWS China).
