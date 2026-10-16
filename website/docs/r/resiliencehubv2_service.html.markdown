---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_service"
description: |-
  Terraform resource for managing an AWS Resilience Hub V2 Service.
---

# Resource: aws_resiliencehubv2_service

Terraform resource for managing an AWS Resilience Hub V2 Service.

A service is the primary building block in Resilience Hub. It comprises AWS resources, code, and observability that together deliver a specific capability. Services can be associated with a resilience policy and a permission model for resource discovery.

## Example Usage

### Basic Usage

```terraform
resource "aws_resiliencehubv2_service" "example" {
  name    = "example-service"
  regions = ["us-west-2"]

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}
```

### With Policy

```terraform
resource "aws_resiliencehubv2_policy" "example" {
  name = "example-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "example" {
  name        = "example-service"
  description = "Production API service"
  policy_arn  = aws_resiliencehubv2_policy.example.arn
  regions     = ["us-west-2", "us-east-1"]

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }

  tags = {
    Environment = "production"
  }
}
```

### With Associated Systems

```terraform
resource "aws_resiliencehubv2_system" "example" {
  name = "example-system"
}

resource "aws_resiliencehubv2_service" "example" {
  name    = "example-service"
  regions = ["us-west-2"]

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }

  associated_system {
    system_arn = aws_resiliencehubv2_system.example.arn
  }
}
```

### With AWS DevOps Agent Integration

```terraform
resource "aws_resiliencehubv2_service" "example" {
  name    = "example-service"
  regions = ["us-west-2"]

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }

  dev_ops_agent_configuration {
    agent_space_arn = "arn:aws:aidevops:us-west-2:123456789012:agentspace/example-space"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the service. Changing this value requires creating a new resource.
* `permission_model` - (Required) Permission model for resource discovery. See [`permission_model` Block](#permission_model-block) below.
* `regions` - (Required) List of AWS regions where the service operates.

The following arguments are optional:

* `associated_system` - (Optional) Systems to associate with the service. See [`associated_system` Block](#associated_system-block) below.
* `dependency_discovery` - (Optional) Dependency discovery. Valid values: `ENABLED`, `DISABLED`.
* `description` - (Optional) Description of the service.
* `dev_ops_agent_configuration` - (Optional) Configuration that links the service to an AWS DevOps Agent Agent Space. See [`dev_ops_agent_configuration` Block](#dev_ops_agent_configuration-block) below. Removing the block disconnects the service from the Agent Space; after a disconnect, `dev_ops_agent_status` is reported as `DISCONNECTED`.
* `kms_key_id` - (Optional) KMS key ARN.
* `policy_arn` - (Optional) ARN of the resilience policy to associate with this service.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `associated_system` Block

The `associated_system` block supports:

* `system_arn` - (Required) ARN of the system to associate with the service.
* `user_journey_ids` - (Optional) List of user journey identifiers that associate the system with the service.

### `dev_ops_agent_configuration` Block

The `dev_ops_agent_configuration` block supports:

* `agent_space_arn` - (Optional) ARN of the AWS DevOps Agent Agent Space to link to this service.

### `permission_model` Block

The `permission_model` block supports:

* `cross_account_role` - (Optional) Cross-account IAM role. See [`cross_account_role` Block](#cross_account_role-block) below.
* `invoker_role_name` - (Required) Name of the IAM role that Resilience Hub assumes for resource discovery.

### `cross_account_role` Block

The `cross_account_role` block supports:

* `cross_account_role_arn` - (Required) ARN of the IAM Role for the profile.
* `external_id` - (Optional) External ID used for assuming the cross-account role.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the service.
* `dev_ops_agent_connected_at` - Timestamp (RFC 3339) when the service successfully connected to the Agent Space.
* `dev_ops_agent_operator_app_url` - URL of the DevOps Agent Operator App for the linked Agent Space. Present when `dev_ops_agent_status` is `CONNECTED`.
* `dev_ops_agent_region` - AWS Region of the linked Agent Space.
* `dev_ops_agent_space_id` - Identifier of the linked Agent Space.
* `dev_ops_agent_status` - Connection status of the AWS DevOps Agent integration. Valid values: `CONNECTING`, `CONNECTED`, `DISCONNECTED`, `FAILED`. Linking is asynchronous: a newly linked service reports `CONNECTING` and transitions to `CONNECTED` (or `FAILED`) out-of-band, so a subsequent refresh is needed to observe the terminal status. After a disconnect, this is reported as `DISCONNECTED`.
* `dev_ops_agent_status_reason` - Human-readable explanation of the current connection status. Present when `dev_ops_agent_status` is `FAILED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_resiliencehubv2_service.example
  identity = {
    arn = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123"
  }
}

resource "aws_resiliencehubv2_service" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the Resilience Hub V2 Service.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub V2 Service using the `arn`. For example:

```terraform
import {
  to = aws_resiliencehubv2_service.example
  id = "arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123"
}
```

Using `terraform import`, import Resilience Hub V2 Service using the `arn`. For example:

```console
% terraform import aws_resiliencehubv2_service.example arn:aws:resiliencehub:us-west-2:123456789012:service/example-service:abc123
```
