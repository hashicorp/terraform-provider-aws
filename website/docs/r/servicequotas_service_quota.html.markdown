---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_service_quota"
description: |-
  Manages an individual Service Quota
---

# Resource: aws_servicequotas_service_quota

Manages an individual Service Quota.

~> **NOTE:** Global quotas apply to all AWS regions, but can only be accessed in `us-east-1` in the Commercial partition or `us-gov-west-1` in the GovCloud partition. In other regions, the AWS API will return the error `The request failed because the specified service does not exist.`

## Example Usage

```terraform
resource "aws_servicequotas_service_quota" "example" {
  quota_code   = "L-F678F1CE"
  service_code = "vpc"
  value        = 75
}
```

## Argument Reference

This resource supports the following arguments:

* `quota_code` - (Optional) Quota code within the service. When configured, the data source directly looks up the service quota. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html). One of `quota_code` or `quota_name` must be specified.
* `quota_name` - (Optional) Quota name within the service. When configured, the data source searches through all service quotas to find the matching quota name. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html). One of `quota_name` or `quota_code` must be specified.
* `service_code` - (Required) Code of the service to track. For example: `vpc`. Available values can be found with the [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).
* `value` - (Required) Float specifying the desired value for the service quota. If the desired value is higher than the current value, a quota increase request is submitted. When a known request is submitted and pending, the value reflects the desired value of the pending request.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `adjustable` - Whether the service quota can be increased.
* `arn` - Amazon Resource Name (ARN) of the service quota.
* `default_value` - Default value of the service quota.
* `id` - Service code and quota code, separated by a front slash (`/`)
* `quota_name` - Name of the quota.
* `service_name` - Name of the service.
* `usage_metric` - Information about the measurement.
    * `metric_dimensions` - The metric dimensions.
        * `class`
        * `resource`
        * `service`
        * `type`
    * `metric_name` - The name of the metric.
    * `metric_namespace` - The namespace of the metric.
    * `metric_statistic_recommendation` - The metric statistic that AWS recommend you use when determining quota usage.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_servicequotas_service_quota` using the service code and quota code, separated by a front slash (`/`). For example:

~> **NOTE:** This resource does not require explicit import and will assume management of an existing service quota on Terraform resource creation.

```terraform
import {
  to = aws_servicequotas_service_quota.example
  id = "vpc/L-F678F1CE"
}
```

Using `terraform import`, import `aws_servicequotas_service_quota` using the service code and quota code, separated by a front slash (`/`). For example:

~> **NOTE:** This resource does not require explicit import and will assume management of an existing service quota on Terraform resource creation.

```console
% terraform import aws_servicequotas_service_quota.example vpc/L-F678F1CE
```
