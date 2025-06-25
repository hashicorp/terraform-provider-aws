---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector2_filter"
description: |-
  Terraform resource for managing an AWS Inspector Filter.
---

# Resource: aws_inspector2_filter

Terraform resource for managing an AWS Inspector Filter.

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_filter" "example" {
  name   = "test"
  action = "NONE"
  filter_criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "111222333444"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `action` - (Required) Action to be applied to the findings that maatch the filter. Possible values are `NONE` and `SUPPRESS`
* `name` - (Required) Name of the filter.
* `filter_criteria` - (Required) Details on the filter criteria. [Documented below](#filter-criteria).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description
* `reason` - (Optional) Reason for creating the filter
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Filter.

### Filter Criteria

The `filter_criteria` configuration block supports the following attributes:

* `aws_account_id` - (Optional) The AWS account ID in which the finding was generated. [Documented below](#string-filter).
* `code_vulnerability_detector_name` - (Optional) The ID of the component. [Documented below](#string-filter).
* `code_vulnerability_detector_tags` - (Optional) The ID of the component. [Documented below](#string-filter).
* `code_vulnerability_file_path` - (Optional) The ID of the component. [Documented below](#string-filter).
* `component_id` - (Optional) The ID of the component. [Documented below](#string-filter).
* `component_type` - (Optional) The type of the component. [Documented below](#string-filter).
* `ec2_instance_image_id` - (Optional) The ID of the Amazon Machine Image (AMI). [Documented below](#string-filter).
* `ec2_instance_subnet_id` - (Optional) The ID of the subnet. [Documented below](#string-filter).
* `ec2_instance_vpc_id` - (Optional) The ID of the VPC. [Documented below](#string-filter).
* `ecr_image_architecture` - (Optional) The architecture of the ECR image. [Documented below](#string-filter).
* `ecr_image_hash` - (Optional) The SHA256 hash of the ECR image. [Documented below](#string-filter).
* `ecr_image_pushed_at` - (Optional) The date range when the image was pushed. [Documented below](#date-filter).
* `ecr_image_registry` - (Optional) The registry of the ECR image. [Documented below](#string-filter).
* `ecr_image_repository_name` - (Optional) The name of the ECR repository. [Documented below](#string-filter).
* `ecr_image_tags` - (Optional) The tags associated with the ECR image. [Documented below](#string-filter).
* `epss_score` - (Optional) EPSS (Exploit Prediction Scoring System) Score of the finding. [Documented below](#number-filter).
* `exploit_available` - (Optional) Availability of exploits. [Documented below](#string-filter).
* `finding_arn` - (Optional) The ARN of the finding. [Documented below](#string-filter).
* `finding_status` - (Optional) The status of the finding. [Documented below](#string-filter).
* `finding_type` - (Optional) The type of the finding. [Documented below](#string-filter).
* `fix_available` - (Optional) Availability of the fix. [Documented below](#string-filter).
* `first_observed_at` - (Optional) When the finding was first observed. [Documented below](#date-filter).
* `inspector_score` - (Optional) The Inspector score given to the finding. [Documented below](#number-filter).
* `lambda_function_execution_role_arn` - (Optional) Lambda execution role ARN. [Documented below](#string-filter).
* `lambda_function_last_modified_at` - (Optional) Last modified timestamp of the lambda function. [Documented below](#date-filter).
* `lambda_function_layers` - (Optional) Lambda function layers. [Documented below](#string-filter).
* `lambda_function_name` - (Optional) Lambda function name. [Documented below](#string-filter).
* `lambda_function_runtime` - (Optional) Lambda function runtime. [Documented below](#string-filter).
* `last_observed_at` - (Optional) When the finding was last observed. [Documented below](#date-filter).
* `network_protocol` - (Optional) The network protocol of the finding. [Documented below](#string-filter).
* `port_range` - (Optional) The port range of the finding. [Documented below](#port-range-filter).
* `related_vulnerabilities` - (Optional) Related vulnerabilities. [Documented below](#string-filter).
* `resource_id` - (Optional) The ID of the resource. [Documented below](#string-filter).
* `resource_tags` - (Optional) The tags of the resource. [Documented below](#map-filter).
* `resource_type` - (Optional) The type of the resource. [Documented below](#string-filter).
* `severity` - (Optional) The severity of the finding. [Documented below](#string-filter).
* `title` - (Optional) The title of the finding. [Documented below](#string-filter).
* `updated_at` - (Optional) When the finding was last updated. [Documented below](#date-filter).
* `vendor_severity` - (Optional) The severity as reported by the vendor. [Documented below](#string-filter).
* `vulnerability_id` - (Optional) The ID of the vulnerability. [Documented below](#string-filter).
* `vulnerability_source` - (Optional) The source of the vulnerability. [Documented below](#string-filter).
* `vulnerable_packages` - (Optional) Details about vulnerable packages. [Documented below](#package-filter).

### Package Filter

The vulnerable package filter configuration block supports the following arguments:

* `architecture` - (Optional) The architecture of the package. [Documented below](#string-filter).
* `epoch` - (Optional) The epoch of the package. [Documented below](#number-filter).
* `file_path` - (Optional) The name of the package. [Documented below](#string-filter).
* `name` - (Optional) The name of the package. [Documented below](#string-filter).
* `release` - (Optional) The release of the package. [Documented below](#string-filter).
* `source_lambda_layer_arn` - (Optional) The ARN of the package's source lambda layer. [Documented below](#string-filter).
* `source_layer_hash` - (Optional) The source layer hash of the package. [Documented below](#string-filter).
* `version` - (Optional) The version of the package. [Documented below](#string-filter).

### String Filter

The string filter configuration block supports the following arguments:

* `comparison` - (Required) The comparison operator. Valid values: `EQUALS`, `PREFIX`, `NOT_EQUALS`.
* `value` - (Required) The value to compare against.

### Number Filter

The number filter configuration block supports the following arguments:

* `lower_inclusive` - (Optional) Lower bound of the range, inclusive.
* `upper_inclusive` - (Optional) Upper bound of the range, inclusive.

### Date Filter

The date filter configuration block supports the following arguments:

* `start_inclusive` - (Optional) Start of the date range in RFC 3339 format, inclusive. Set the timezone to UTC.
* `end_inclusive` - (Optional) End of the date range in RFC 3339 format, inclusive. Set the timezone to UTC.

### Map Filter

The map filter configuration block supports the following arguments:

* `comparison` - (Required) The comparison operator. Valid values: `EQUALS`.
* `key` - (Required) The key to filter on.
* `value` - (Required) The value to filter on.

### Port Range Filter

The port range filter configuration block supports the following arguments:

* `begin_inclusive` - (Required) The beginning of the port range, inclusive.
* `end_inclusive` - (Required) The end of the port range, inclusive.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Inspector Filter using the `arn`. For example:

```terraform
import {
  to = aws_inspector2_filter.example
  id = "arn:aws:inspector2:us-east-1:111222333444:owner/111222333444/filter/abcdefgh12345678"
}
```

Using `terraform import`, import Inspector Filter using the `example_id_arg`. For example:

```console
% terraform import aws_inspector2_filter.example "arn:aws:inspector2:us-east-1:111222333444:owner/111222333444/filter/abcdefgh12345678"
```
