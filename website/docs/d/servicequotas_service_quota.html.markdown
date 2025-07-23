---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_service_quota"
description: |-
  Retrieve information about a Service Quota
---

# Data Source: aws_servicequotas_service_quota

Retrieve information about a Service Quota.

~> **NOTE:** Global quotas apply to all AWS regions, but can only be accessed in `us-east-1` in the Commercial partition or `us-gov-west-1` in the GovCloud partition. In other regions, the AWS API will return the error `The request failed because the specified service does not exist.`

## Example Usage

```terraform
data "aws_servicequotas_service_quota" "by_quota_code" {
  quota_code   = "L-F678F1CE"
  service_code = "vpc"
}

data "aws_servicequotas_service_quota" "by_quota_name" {
  quota_name   = "VPCs per Region"
  service_code = "vpc"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_code` - (Required) Service code for the quota. Available values can be found with the [`aws_servicequotas_service` data source](/docs/providers/aws/d/servicequotas_service.html) or [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).
* `quota_code` - (Optional) Quota code within the service. When configured, the data source directly looks up the service quota. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html). One of `quota_code` or `quota_name` must be specified.
* `quota_name` - (Optional) Quota name within the service. When configured, the data source searches through all service quotas to find the matching quota name. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html). One of `quota_name` or `quota_code` must be specified.

~> *NOTE:* Either `quota_code` or `quota_name` must be configured.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `adjustable` - Whether the service quota is adjustable.
* `arn` - ARN of the service quota.
* `default_value` - Default value of the service quota.
* `global_quota` - Whether the service quota is global for the AWS account.
* `id` - ARN of the service quota.
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
* `value` - Current value of the service quota.
