---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_container_service"
description: |- 
  Manages a Lightsail container service for running containerized applications.
---

# Resource: aws_lightsail_container_service

Manages a Lightsail container service. Use this resource to create and manage a scalable compute and networking platform for deploying, running, and managing containerized applications in Lightsail.

~> **Note:** For more information about the AWS Regions in which you can create Amazon Lightsail container services, see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail).

~> **NOTE:** You must create and validate an SSL/TLS certificate before you can use `public_domain_names` with your container service. For more information, see [Enabling and managing custom domains for your Amazon Lightsail container services](https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-creating-container-services-certificates).

## Example Usage

### Basic Usage

```terraform
resource "aws_lightsail_container_service" "example" {
  name        = "container-service-1"
  power       = "nano"
  scale       = 1
  is_disabled = false

  tags = {
    foo1 = "bar1"
    foo2 = ""
  }
}
```

### Public Domain Names

```terraform
resource "aws_lightsail_container_service" "example" {
  # ... other configuration ...

  public_domain_names {
    certificate {
      certificate_name = "example-certificate"
      domain_names = [
        "www.example.com",
        # maybe another domain name
      ]
    }

    # certificate {
    # maybe another certificate
    # }
  }
}
```

### Private Registry Access

```terraform
resource "aws_lightsail_container_service" "example" {
  # ... other configuration ...

  private_registry_access {
    ecr_image_puller_role {
      is_active = true
    }
  }
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = [aws_lightsail_container_service.example.private_registry_access[0].ecr_image_puller_role[0].principal_arn]
    }

    actions = [
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer",
    ]
  }
}

resource "aws_ecr_repository_policy" "example" {
  repository = aws_ecr_repository.example.name
  policy     = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the container service. Names must be of length 1 to 63, and be unique within each AWS Region in your Lightsail account.
* `power` - (Required) Power specification for the container service. The power specifies the amount of memory, the number of vCPUs, and the monthly price of each node of the container service. Possible values: `nano`, `micro`, `small`, `medium`, `large`, `xlarge`.
* `scale` - (Required) Scale specification for the container service. The scale specifies the allocated compute nodes of the container service.

The following arguments are optional:

* `is_disabled` - (Optional) Whether to disable the container service. Defaults to `false`.
* `private_registry_access` - (Optional) Configuration for the container service to access private container image repositories, such as Amazon Elastic Container Registry (Amazon ECR) private repositories. [See below](#private-registry-access).
* `public_domain_names` - (Optional) Public domain names to use with the container service, such as example.com and www.example.com. You can specify up to four public domain names for a container service. The domain names that you specify are used when you create a deployment with a container configured as the public endpoint of your container service. If you don't specify public domain names, then you can use the default domain of the container service. [See below](#public-domain-names).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

### Private Registry Access

The `private_registry_access` block supports the following arguments:

* `ecr_image_puller_role` - (Optional) Configuration to access private container image repositories, such as Amazon Elastic Container Registry (Amazon ECR) private repositories. [See below](#ecr-image-puller-role).

### ECR Image Puller Role

The `ecr_image_puller_role` block supports the following arguments:

* `is_active` - (Optional) Whether to activate the role. Defaults to `false`.

### Public Domain Names

The `public_domain_names` block supports the following arguments:

* `certificate` - (Required) Set of certificate configurations for the public domain names. Each element contains the following attributes:
    * `certificate_name` - (Required) Name of the certificate.
    * `domain_names` - (Required) List of domain names for the certificate.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the container service.
* `availability_zone` - Availability Zone. Follows the format us-east-2a (case-sensitive).
* `created_at` - Date and time when the container service was created.
* `id` - Same as `name`.
* `power_id` - Power ID of the container service.
* `principal_arn` - Principal ARN of the container service. The principal ARN can be used to create a trust relationship between your standard AWS account and your Lightsail container service. This allows you to give your service permission to access resources in your standard AWS account.
* `private_domain_name` - Private domain name of the container service. The private domain name is accessible only by other resources within the default virtual private cloud (VPC) of your Lightsail account.
* `private_registry_access` - Configuration for the container service to access private container image repositories. Contains the following attributes:
    * `ecr_image_puller_role` - Configuration to access private container image repositories. Contains the following attributes:
        * `principal_arn` - Principal ARN of the container service. The principal ARN can be used to create a trust relationship between your standard AWS account and your Lightsail container service.
* `region_name` - AWS Region name.
* `resource_type` - Lightsail resource type of the container service (i.e., ContainerService).
* `state` - Current state of the container service.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.
* `url` - Publicly accessible URL of the container service. If no public endpoint is specified in the currentDeployment, this URL returns a 404 response.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lightsail Container Service using the `name`. For example:

```terraform
import {
  to = aws_lightsail_container_service.example
  id = "container-service-1"
}
```

Using `terraform import`, import Lightsail Container Service using the `name`. For example:

```console
% terraform import aws_lightsail_container_service.example container-service-1
```
