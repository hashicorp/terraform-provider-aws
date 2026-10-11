---
subcategory: "Lambda MicroVMs"
layout: "aws"
page_title: "AWS: aws_lambdamicrovms_image_version"
description: |-
  Provides details about an AWS Lambda MicroVMs Image Version.
---

# Data Source: aws_lambdamicrovms_image_version

Provides details about an AWS Lambda MicroVMs Image Version.

## Example Usage

### Basic Usage

```terraform
data "aws_lambdamicrovms_image_version" "example" {
  image_identifier = aws_lambdamicrovms_image.example.arn
  image_version    = aws_lambdamicrovms_image.example.image_version
}
```

## Argument Reference

The following arguments are required:

* `image_identifier` - (Required) ARN or ID of the MicroVM image.
* `image_version` - (Required) Version of the MicroVM image.

The following arguments are optional:

* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `additional_os_capabilities` - Additional OS capabilities granted to the MicroVM runtime environment.
* `base_image_arn` - ARN of the base MicroVM image used.
* `base_image_version` - Version of the base MicroVM image used.
* `build_role_arn` - ARN of the IAM role used to build the image version.
* `code_artifact` - Code artifact for this version. See [`code_artifact`](#code_artifact) below.
* `cpu_configuration` - List of supported CPU configurations. See [`cpu_configuration`](#cpu_configuration) below.
* `created_at` - Timestamp when the version was created.
* `description` - Description of the version.
* `egress_network_connectors` - List of egress network connectors available to the MicroVM at runtime.
* `environment_variables` - Environment variables set in the MicroVM runtime environment.
* `hooks` - Lifecycle hook configuration. See [`hooks`](#hooks) below.
* `image_arn` - ARN of the MicroVM image.
* `logging` - Logging configuration for this version. See [`logging`](#logging) below.
* `resources` - Resource requirements for the MicroVM. See [`resources`](#resources) below.
* `state` - Current state of the version.
* `state_reason` - Reason for the current state.
* `status` - Availability status of the version. `ACTIVE` versions can launch new MicroVMs; `INACTIVE` versions cannot.
* `tags` - Map of tags assigned to the version.
* `updated_at` - Timestamp when the version was last updated.

### `code_artifact`

* `uri` - URI of the code artifact.

### `cpu_configuration`

* `architecture` - CPU architecture.

### `hooks`

* `microvm_hooks` - Lifecycle hooks for MicroVM events. See [`microvm_hooks`](#microvm_hooks) below.
* `microvm_image_hooks` - Hooks for MicroVM image build events. See [`microvm_image_hooks`](#microvm_image_hooks) below.
* `port` - Port number on which the hooks listener runs.

### `microvm_hooks`

* `resume` - Whether the resume hook is `ENABLED` or `DISABLED`.
* `resume_timeout_in_seconds` - Maximum time in seconds for the resume hook to complete.
* `run` - Whether the run hook is `ENABLED` or `DISABLED`.
* `run_timeout_in_seconds` - Maximum time in seconds for the run hook to complete.
* `suspend` - Whether the suspend hook is `ENABLED` or `DISABLED`.
* `suspend_timeout_in_seconds` - Maximum time in seconds for the suspend hook to complete.
* `terminate` - Whether the terminate hook is `ENABLED` or `DISABLED`.
* `terminate_timeout_in_seconds` - Maximum time in seconds for the terminate hook to complete.

### `microvm_image_hooks`

* `ready` - Whether the ready hook is `ENABLED` or `DISABLED`.
* `ready_timeout_in_seconds` - Maximum time in seconds for the ready hook to complete.
* `validate` - Whether the validate hook is `ENABLED` or `DISABLED`.
* `validate_timeout_in_seconds` - Maximum time in seconds for the validate hook to complete.

### `logging`

* `cloudwatch` - CloudWatch Logs configuration. See [`cloudwatch`](#cloudwatch) below.
* `disabled` - Present when logging is disabled.

### `cloudwatch`

* `log_group` - Name of the CloudWatch Logs log group.
* `log_stream` - Name of the CloudWatch Logs log stream.

### `resources`

* `minimum_memory_in_mib` - Minimum amount of memory in MiB allocated to the MicroVM.
