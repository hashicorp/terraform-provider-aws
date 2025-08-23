---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_infrastructure_configuration"
description: |-
    Provides details about an Image Builder Infrastructure Configuration
---

# Data Source: aws_imagebuilder_infrastructure_configuration

Provides details about an Image Builder Infrastructure Configuration.

## Example Usage

```terraform
data "aws_imagebuilder_infrastructure_configuration" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:infrastructure-configuration/example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Required) ARN of the infrastructure configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `date_created` - Date the infrastructure configuration was created.
* `date_created` - Date the infrastructure configuration was updated.
* `description` - Description of the infrastructure configuration.
* `instance_metadata_options` - Nested list of instance metadata options for the HTTP requests that pipeline builds use to launch EC2 build and test instances.
    * `http_put_response_hop_limit` - Number of hops that an instance can traverse to reach its destonation.
    * `http_tokens` - Whether a signed token is required for instance metadata retrieval requests.
* `instance_profile_name` - Name of the IAM Instance Profile associated with the configuration.
* `instance_types` - Set of EC2 Instance Types associated with the configuration.
* `key_pair` - Name of the EC2 Key Pair associated with the configuration.
* `logging` - Nested list of logging settings.
    * `s3_logs` - Nested list of S3 logs settings.
        * `s3_bucket_name` - Name of the S3 Bucket for logging.
        * `s3_key_prefix` - Key prefix for S3 Bucket logging.
* `name` - Name of the infrastructure configuration.
* `placement` - Placement settings that define where the instances that are launched from your image will run.
    * `availability_zone` - Availability Zone where your build and test instances will launch.
    * `host_id` - ID of the Dedicated Host on which build and test instances run.
    * `host_resource_group_arn` - ARN of the host resource group in which to launch build and test instances.
    * `tenancy` - Placement tenancy of the instance.
* `resource_tags` - Key-value map of resource tags for the infrastructure created by the infrastructure configuration.
* `security_group_ids` - Set of EC2 Security Group identifiers associated with the configuration.
* `sns_topic_arn` - ARN of the SNS Topic associated with the configuration.
* `subnet_id` - Identifier of the EC2 Subnet associated with the configuration.
* `tags` - Key-value map of resource tags for the infrastructure configuration.
* `terminate_instance_on_failure` - Whether instances are terminated on failure.
