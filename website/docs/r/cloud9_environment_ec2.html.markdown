---
subcategory: "Cloud9"
layout: "aws"
page_title: "AWS: aws_cloud9_environment_ec2"
description: |-
  Provides a Cloud9 EC2 Development Environment.
---

# Resource: aws_cloud9_environment_ec2

Provides a Cloud9 EC2 Development Environment.

## Example Usage

Basic usage:

```terraform
resource "aws_cloud9_environment_ec2" "example" {
  instance_type = "t2.micro"
  name          = "example-env"
  image_id      = "amazonlinux-2023-x86_64"
}
```

Get the URL of the Cloud9 environment after creation:

```terraform
resource "aws_cloud9_environment_ec2" "example" {
  instance_type = "t2.micro"
}

data "aws_instance" "cloud9_instance" {
  filter {
    name = "tag:aws:cloud9:environment"
    values = [
    aws_cloud9_environment_ec2.example.id]
  }
}

output "cloud9_url" {
  value = "https://${var.region}.console.aws.amazon.com/cloud9/ide/${aws_cloud9_environment_ec2.example.id}"
}
```

Allocate a static IP to the Cloud9 environment:

```terraform
resource "aws_cloud9_environment_ec2" "example" {
  instance_type = "t2.micro"
}

data "aws_instance" "cloud9_instance" {
  filter {
    name = "tag:aws:cloud9:environment"
    values = [
    aws_cloud9_environment_ec2.example.id]
  }
}

resource "aws_eip" "cloud9_eip" {
  instance = data.aws_instance.cloud9_instance.id
  domain   = "vpc"
}

output "cloud9_public_ip" {
  value = aws_eip.cloud9_eip.public_ip
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the environment.
* `instance_type` - (Required) The type of instance to connect to the environment, e.g., `t2.micro`.
* `image_id` - (Required) The identifier for the Amazon Machine Image (AMI) that's used to create the EC2 instance. Valid values are
    * `amazonlinux-2-x86_64`
    * `amazonlinux-2023-x86_64`
    * `ubuntu-18.04-x86_64`
    * `ubuntu-22.04-x86_64`
    * `resolve:ssm:/aws/service/cloud9/amis/amazonlinux-2-x86_64`
    * `resolve:ssm:/aws/service/cloud9/amis/amazonlinux-2023-x86_64`
    * `resolve:ssm:/aws/service/cloud9/amis/ubuntu-18.04-x86_64`
    * `resolve:ssm:/aws/service/cloud9/amis/ubuntu-22.04-x86_64`
* `automatic_stop_time_minutes` - (Optional) The number of minutes until the running instance is shut down after the environment has last been used.
* `connection_type` - (Optional) The connection type used for connecting to an Amazon EC2 environment. Valid values are `CONNECT_SSH` and `CONNECT_SSM`. For more information please refer [AWS documentation for Cloud9](https://docs.aws.amazon.com/cloud9/latest/user-guide/ec2-ssm.html).
* `description` - (Optional) The description of the environment.
* `owner_arn` - (Optional) The ARN of the environment owner. This can be ARN of any AWS IAM principal. Defaults to the environment's creator.
* `subnet_id` - (Optional) The ID of the subnet in Amazon VPC that AWS Cloud9 will use to communicate with the Amazon EC2 instance.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the environment.
* `arn` - The ARN of the environment.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `type` - The type of the environment (e.g., `ssh` or `ec2`).
