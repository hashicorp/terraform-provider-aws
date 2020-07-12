---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_infrastructure_configuration"
description: |-
  Provides AWS Image Builder infrasctracture configuration 
---

# Resource: aws_imagebuilder_infrastructure_configuration
Provides AWS Image Builder infrasctracture configuration

## Example Usage

```hcl
# Create new infrastructure configuration
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test_profile.name
  name                  = "example"
  description           = "example desc"
  instance_types        = ["t2.nano", "t3.micro"]
  key_pair              = aws_key_pair.example.key_name
  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.example.bucket
      s3_key_prefix  = "logs"
    }
  }
  security_group_ids            = [aws_security_group.example.id]
  sns_topic_arn                 = aws_sns_topic.example.arn
  subnet_id                     = aws_subnet.main.id
  terminate_instance_on_failure = true
  tags = {
    foo = "bar"
  }
}
```

## Argument Reference

The following arguments are supported:

* `instance_profile_name` - (Required) The name of iam instance profile.
* `name` - (Required) The name of the imgage builder infrastructure configuration.
* `description` - (Optional) The description of the imgage builder infrastructure configuration.
* `instance_types` - (Optional) A list of instance to use.
* `key_pair` - (Optional) The name of your key pair
* `logging` - (Optional) The logging configuration, see below for details.
* `security_group_ids` - (Optional) A list of security groups to assign to resource, mandatory if subnet provider.
* `sns_topic_arn` - (Optional) The ARN of sns topic to use.
* `subnet_id` - (Optional) The id of subnet to use, if provided requires that security groups also be provided. 
* `tags` - (Optional) A map of tags to assign to the resource.
* `terminate_instance_on_failure` - (Optional) should the instance be terminated when the pipeline fails (Default: false).

### Logging
The logging object support the following blocks:
* s3_logs

The s3_logs block supports the following arguments:
* `s3_bucket_name` - (Required) The name of the bucket to use.
* `s3_key_prefix` - (Optional) The prefix to use for logs (Default: "/").

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of the configuration (matches `arn`).
* `arn` - The ARN of the configuration  (matches `id`).
* `date_created` - The timestamp when the configuration was created.
* `date_updated` - The timestamp when the configuration was updated.

## Import

Infrastructure configuration can be imported using the ARN , e.g.

```
$ terraform import aws_imagebuilder_infrastructure_configuration.test '<ARN>'
```
