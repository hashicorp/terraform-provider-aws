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

### Example Usage with Wait For Fulfillment

```terraform
resource "aws_servicequotas_service_quota" "example" {
  quota_code           = "L-F678F1CE"
  service_code         = "vpc"
  value                = 75
  wait_for_fulfillment = true
}
```

### Example Usage with Custom Timeouts

When using `wait_for_fulfillment`, you may want to configure longer timeouts if quota approval typically takes more than the default 10 minutes:

```terraform
resource "aws_servicequotas_service_quota" "example" {
  quota_code           = "L-F678F1CE"
  service_code         = "vpc"
  value                = 75
  wait_for_fulfillment = true

  timeouts {
    create = "30m"
    update = "30m"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `quota_code` - (Required) Code of the service quota to track. For example: `L-F678F1CE`. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html).
* `service_code` - (Required) Code of the service to track. For example: `vpc`. Available values can be found with the [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).
* `value` - (Required) Float specifying the desired value for the service quota. If the desired value is higher than the current value, a quota increase request is submitted. When a known request is submitted and pending, the value reflects the desired value of the pending request.
* `wait_for_fulfillment` - (Optional) Boolean indicating whether the resource should wait for the quota increase request to be fulfilled before completing. Defaults to `false`. When set to `true`, Terraform will wait for the request to move from a pending state to an approved state before marking the resource as successfully created or updated. This is useful for automation scenarios where subsequent resources depend on the increased quota being available.

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

## Timeouts

~> **NOTE:** When using `wait_for_fulfillment = true`, quota increase requests may take longer than the default timeout to be approved and enacted by AWS. Consider configuring longer timeouts if needed.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)

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
