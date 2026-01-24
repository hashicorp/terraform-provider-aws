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

This data source supports the following arguments:

* `service_name` - (Required) Name of the service you want to generate a Service Principal Name for.
* `region` - (Optional) Region you'd like the SPN for. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Identifier of the current Service Principal (compound of service, Region and suffix). (e.g. `logs.us-east-1.amazonaws.com`in AWS Commercial, `logs.cn-north-1.amazonaws.com.cn` in AWS China).
* `name` - Service Principal Name (e.g., `logs.amazonaws.com` in AWS Commercial, `logs.amazonaws.com.cn` in AWS China).
* `service` - Service used for SPN generation (e.g. `logs`).
* `suffix` - Suffix of the SPN (e.g., `amazonaws.com` in AWS Commercial, `amazonaws.com.cn` in AWS China).
