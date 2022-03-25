---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_distribution_configuration"
description: |-
    Provides details about an Image Builder Distribution Configuration
---

# Data Source: aws_imagebuilder_distribution_configuration

Provides details about an Image Builder Distribution Configuration.

## Example Usage

```terraform
data "aws_imagebuilder_distribution_configuration" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:distribution-configuration/example"
}
```

## Argument Reference

* `arn` - (Required) Amazon Resource Name (ARN) of the distribution configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `date_created` - Date the distribution configuration was created.
* `date_updated` - Date the distribution configuration was updated.
* `description` - Description of the distribution configuration.
* `distribution` - Set of distributions.
    * `ami_distribution_configuration` - Nested list of AMI distribution configuration.
        * `ami_tags` - Key-value map of tags to apply to distributed AMI.
        * `description` - Description to apply to distributed AMI.
        * `kms_key_id` - Amazon Resource Name (ARN) of Key Management Service (KMS) Key to encrypt AMI.
        * `launch_permission` - Nested list of EC2 launch permissions.
            * `organization_arns` - Set of AWS Organization ARNs.
            * `organizational_unit_arns` - Set of AWS Organizational Unit ARNs.
            * `user_groups` - Set of EC2 launch permission user groups.
            * `user_ids` - Set of AWS Account identifiers.
        * `target_account_ids` - Set of target AWS Account identifiers.
    * `container_distribution_configuration` - Nested list of container distribution configurations.
        * `container_tags` - Set of tags that are attached to the container distribution configuration.
        * `description` - Description of the container distribution configuration.
        * `target_repository` - Set of destination repositories for the container distribution configuration.
            * `repository_name` - Name of the container repository where the output container image is stored.
            * `service` - Service in which the image is registered.
    * `launch_template_configuration` - Nested list of launch template configurations.
        * `default` - Indicates whether the specified Amazon EC2 launch template is set as the default launch template.
        * `launch_template_id` - ID of the Amazon EC2 launch template.
    * `license_configuration_arns` - Set of Amazon Resource Names (ARNs) of License Manager License Configurations.
    * `region` - AWS Region of distribution.
* `name` - Name of the distribution configuration.
* `tags` - Key-value map of resource tags for the distribution configuration.
