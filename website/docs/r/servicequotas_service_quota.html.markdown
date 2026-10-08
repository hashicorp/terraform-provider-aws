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

When using `wait_for_fulfillment = true`, configure longer timeouts if quota fulfillment typically takes more than the default 10 minutes:

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

* `quota_code` - (Required) Code of the service quota to track. For example: `L-F678F1CE`. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html).
* `service_code` - (Required) Code of the service to track. For example: `vpc`. Available values can be found with the [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).
* `value` - (Required) Desired value for the service quota. If higher than the current value, a quota increase request is submitted unless a suitable request is reused with waiting enabled. With `wait_for_fulfillment = true`, state reflects the applied quota value, or the AWS default value if no applied value is available. With waiting disabled, a known pending request's desired value is reflected in state.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `wait_for_fulfillment` - (Optional) Whether to wait until the applied quota value is greater than or equal to the configured `value` before completing creation or update. Defaults to `false`. Request approval is not a prerequisite for completion.

With waiting enabled, `value` is a minimum target. Creation and updates succeed when the applied quota already meets or exceeds this target. Higher applied values remain in state without causing recurring plan differences. With waiting disabled, configuring a value below the current quota remains unsupported.

With `wait_for_fulfillment = true`, creation and updates reuse suitable pending or open requests. An open request for less than the configured target causes an error instead of submitting a duplicate request. When no open request is found, history recovery selects the newest eligible `APPROVED` or `CASE_CLOSED` request whose `DesiredValue` is at least the configured `value`. `DesiredValue` is request metadata and does not confirm the applied quota value.

If the applied value is below the configured target, `DENIED`, `NOT_APPROVED`, and `INVALID_REQUEST` statuses cause the operation to fail. For `PENDING`, `CASE_OPENED`, `APPROVED`, and `CASE_CLOSED`, polling continues until the applied value meets the target or the timeout expires. `CASE_CLOSED` does not establish approval or denial, and recovering a closed request from history does not establish fulfillment. Unknown request statuses cause an error.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Service code and quota code, separated by a front slash (`/`).
* `adjustable` - Whether the service quota can be increased.
* `arn` - ARN of the service quota.
* `default_value` - Default value of the service quota.
* `quota_name` - Name of the quota.
* `request_id` - ID of the tracked quota increase request. With waiting enabled, retained across timeouts. Refresh may retain approved or closed requests until their full requested value is observed. Successful creation or update clears the ID once the applied quota meets or exceeds the configured `value`.
* `request_status` - Last observed status of the tracked quota increase request. Request status alone does not establish fulfillment.
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

~> **Note:** With `wait_for_fulfillment = true`, the configured `create` or `update` timeout is shared across request discovery, submission, and applied-value polling. If waiting times out after a request is identified, Terraform retains the request association so a subsequent apply resumes waiting for that request. Quota fulfillment may take longer than the default timeout; configure a longer timeout if needed.

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
