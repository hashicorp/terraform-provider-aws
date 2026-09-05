---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_service_quotas"
description: |-
  Retrieve information about all Service Quotas for a service
---

# Data Source: aws_servicequotas_service_quotas

Retrieve information about all Service Quotas for a service.

~> **NOTE:** Global quotas apply to all AWS regions, but can only be accessed in `us-east-1` in the Commercial partition or `us-gov-west-1` in the GovCloud partition. In other regions, the AWS API will return the error `The request failed because the specified service does not exist.`

## Example Usage

```terraform
data "aws_servicequotas_service_quotas" "example" {
  service_code = "vpc"
}
```

## Argument Reference

This data source supports the following arguments:

* `service_code` - (Required) Service code for the quota. Available values can be found with the [`aws_servicequotas_service` data source](/docs/providers/aws/d/servicequotas_service.html) or [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).
* `region` - (Optional) Region associated with the quota. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Service code.
* `quotas` - List of all quotas for the service. Each element contains the following attributes:
    * `adjustable` - Whether the service quota is adjustable.
    * `arn` - ARN of the service quota.
    * `default_value` - Default value of the service quota.
    * `global_quota` - Whether the service quota is global for the AWS account.
    * `quota_code` - Quota code for the service quota.
    * `quota_name` - Name of the service quota.
    * `service_code` - Service code of the service quota.
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
    * `value` - Default value of the service quota.
