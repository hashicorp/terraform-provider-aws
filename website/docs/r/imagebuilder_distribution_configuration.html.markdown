---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_distribution_configuration"
description: |-
    Manage an Image Builder Distribution Configuration
---

# Resource: aws_imagebuilder_distribution_configuration

Manages an Image Builder Distribution Configuration.

## Example Usage

```terraform
resource "aws_imagebuilder_distribution_configuration" "example" {
  name = "example"

  distribution {
    ami_distribution_configuration {
      ami_tags = {
        CostCenter = "IT"
      }

      name = "example-{{ imagebuilder:buildDate }}"

      launch_permission {
        user_ids = ["123456789012"]
      }
    }

    launch_template_configuration {
      launch_template_id = "lt-0aaa1bcde2ff3456"
    }

    region = "us-east-1"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the distribution configuration.
* `distribution` - (Required) One or more configuration blocks with distribution settings. Detailed below.

The following arguments are optional:

* `description` - (Optional) Description of the distribution configuration.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key used to encrypt the distribution configuration.
* `tags` - (Optional) Key-value map of resource tags for the distribution configuration. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### distribution

The following arguments are required:

* `region` - (Required) AWS Region for the distribution.

The following arguments are optional:

* `ami_distribution_configuration` - (Optional) Configuration block with Amazon Machine Image (AMI) distribution settings. Detailed below.
* `container_distribution_configuration` - (Optional) Configuration block with container distribution settings. Detailed below.
* `launch_template_configuration` - (Optional) Set of launch template configuration settings that apply to image distribution. Detailed below.
* `license_configuration_arns` - (Optional) Set of Amazon Resource Names (ARNs) of License Manager License Configurations.

### ami_distribution_configuration

The following arguments are optional:

* `ami_tags` - (Optional) Key-value map of tags to apply to the distributed AMI.
* `description` - (Optional) Description to apply to the distributed AMI.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key to encrypt the distributed AMI.
* `launch_permission` - (Optional) Configuration block of EC2 launch permissions to apply to the distributed AMI. Detailed below.
* `name` - (Optional) Name to apply to the distributed AMI.
* `target_account_ids` - (Optional) Set of AWS Account identifiers to distribute the AMI.

### launch_permission

The following arguments are optional:

* `organization_arns` - (Optional) Set of AWS Organization ARNs to assign.
* `organizational_unit_arns` - (Optional) Set of AWS Organizational Unit ARNs to assign.
* `user_groups` - (Optional) Set of EC2 launch permission user groups to assign. Use `all` to distribute a public AMI.
* `user_ids` - (Optional) Set of AWS Account identifiers to assign.

### container_distribution_configuration

* `container_tags` - (Optional) Set of tags that are attached to the container distribution configuration.
* `description` - (Optional) Description of the container distribution configuration.
* `target_repository` (Required) Configuration block with the destination repository for the container distribution configuration.

### target_repository

* `repository_name` - (Required) The name of the container repository where the output container image is stored. This name is prefixed by the repository location.
* `service` - (Required) The service in which this image is registered. Valid values: `ECR`.

### launch_template_configuration

* `default` - (Optional) Indicates whether to set the specified Amazon EC2 launch template as the default launch template. Defaults to `true`.
* `account_id` - The account ID that this configuration applies to.
* `launch_template_id` - (Required) The ID of the Amazon EC2 launch template to use.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - (Required) Amazon Resource Name (ARN) of the distribution configuration.
* `date_created` - Date the distribution configuration was created.
* `date_updated` - Date the distribution configuration was updated.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_imagebuilder_distribution_configurations` resources can be imported by using the Amazon Resource Name (ARN), e.g.,

```
$ terraform import aws_imagebuilder_distribution_configuration.example arn:aws:imagebuilder:us-east-1:123456789012:distribution-configuration/example
```
