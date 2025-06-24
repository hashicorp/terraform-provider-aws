---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_infrastructure_configuration"
description: |-
  Manages an Image Builder Infrastructure Configuration
---

# Resource: aws_imagebuilder_infrastructure_configuration

Manages an Image Builder Infrastructure Configuration.

## Example Usage

```terraform
resource "aws_imagebuilder_infrastructure_configuration" "example" {
  description                   = "example description"
  instance_profile_name         = aws_iam_instance_profile.example.name
  instance_types                = ["t2.nano", "t3.micro"]
  key_pair                      = aws_key_pair.example.key_name
  name                          = "example"
  security_group_ids            = [aws_security_group.example.id]
  sns_topic_arn                 = aws_sns_topic.example.arn
  subnet_id                     = aws_subnet.main.id
  terminate_instance_on_failure = true

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.example.bucket
      s3_key_prefix  = "logs"
    }
  }

  tags = {
    foo = "bar"
  }
}
```

## Argument Reference

The following arguments are required:

* `instance_profile_name` - (Required) Name of IAM Instance Profile.
* `name` - (Required) Name for the configuration.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description for the configuration.
* `instance_metadata_options` - (Optional) Configuration block with instance metadata options for the HTTP requests that pipeline builds use to launch EC2 build and test instances. Detailed below.
* `instance_types` - (Optional) Set of EC2 Instance Types.
* `key_pair` - (Optional) Name of EC2 Key Pair.
* `logging` - (Optional) Configuration block with logging settings. Detailed below.
* `placement` - (Optional) Configuration block with placement settings that define where the instances that are launched from your image will run. Detailed below.
* `resource_tags` - (Optional) Key-value map of resource tags to assign to infrastructure created by the configuration.
* `security_group_ids` - (Optional) Set of EC2 Security Group identifiers.
* `sns_topic_arn` - (Optional) Amazon Resource Name (ARN) of SNS Topic.
* `subnet_id` - (Optional) EC2 Subnet identifier. Also requires `security_group_ids` argument.
* `tags` - (Optional) Key-value map of resource tags to assign to the configuration. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `terminate_instance_on_failure` - (Optional) Enable if the instance should be terminated when the pipeline fails. Defaults to `false`.

### instance_metadata_options

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `http_put_response_hop_limit` - The number of hops that an instance can traverse to reach its destonation.
* `http_tokens` - Whether a signed token is required for instance metadata retrieval requests. Valid values: `required`, `optional`.

### logging

The following arguments are required:

* `s3_logs` - (Required) Configuration block with S3 logging settings. Detailed below.

### s3_logs

The following arguments are required:

* `s3_bucket_name` - (Required) Name of the S3 Bucket.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `s3_key_prefix` - (Optional) Prefix to use for S3 logs. Defaults to `/`.

### placement

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `availability_zone` - (Optional) Availability Zone where your build and test instances will launch.
* `host_id` - (Optional) ID of the Dedicated Host on which build and test instances run. Conflicts with `host_resource_group_arn`.
* `host_resource_group_arn` - (Optional) ARN of the host resource group in which to launch build and test instances. Conflicts with `host_id`.
* `tenancy` - (Optional) Placement tenancy of the instance. Valid values: `default`, `dedicated` and `host`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the configuration.
* `arn` - Amazon Resource Name (ARN) of the configuration.
* `date_created` - Date when the configuration was created.
* `date_updated` - Date when the configuration was updated.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_imagebuilder_infrastructure_configuration` using the Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_imagebuilder_infrastructure_configuration.example
  id = "arn:aws:imagebuilder:us-east-1:123456789012:infrastructure-configuration/example"
}
```

Using `terraform import`, import `aws_imagebuilder_infrastructure_configuration` using the Amazon Resource Name (ARN). For example:

```console
% terraform import aws_imagebuilder_infrastructure_configuration.example arn:aws:imagebuilder:us-east-1:123456789012:infrastructure-configuration/example
```
