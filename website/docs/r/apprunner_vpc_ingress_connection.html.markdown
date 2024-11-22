---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_vpc_ingress_connection"
description: |-
  Manages an App Runner VPC Ingress Connection.
---

# Resource: aws_apprunner_vpc_ingress_connection

Manages an App Runner VPC Ingress Connection.

## Example Usage

```terraform
resource "aws_apprunner_vpc_ingress_connection" "example" {
  name        = "example"
  service_arn = aws_apprunner_service.example.arn

  ingress_vpc_configuration {
    vpc_id          = aws_default_vpc.default.id
    vpc_endpoint_id = aws_vpc_endpoint.apprunner.id
  }

  tags = {
    foo = "bar"
  }
}

```

## Argument Reference

The following arguments supported:

* `name` - (Required) A name for the VPC Ingress Connection resource. It must be unique across all the active VPC Ingress Connections in your AWS account in the AWS Region.
* `service_arn` - (Required) The Amazon Resource Name (ARN) for this App Runner service that is used to create the VPC Ingress Connection resource.
* `ingress_vpc_configuration` - (Required) Specifications for the customer’s Amazon VPC and the related AWS PrivateLink VPC endpoint that are used to create the VPC Ingress Connection resource. See [Ingress VPC Configuration](#ingress-vpc-configuration) below for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Ingress VPC Configuration

The `ingress_vpc_configuration` block supports the following argument:

* `vpc_id` - (Required) The ID of the VPC that is used for the VPC endpoint.
* `vpc_endpoint_id` - (Required) The ID of the VPC endpoint that your App Runner service connects to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the VPC Ingress Connection.
* `domain_name` - The domain name associated with the VPC Ingress Connection resource.
* `status` - The current status of the VPC Ingress Connection.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Runner VPC Ingress Connection using the `arn`. For example:

```terraform
import {
  to = aws_apprunner_vpc_ingress_connection.example
  id = "arn:aws:apprunner:us-west-2:837424938642:vpcingressconnection/example/b379f86381d74825832c2e82080342fa"
}
```

Using `terraform import`, import App Runner VPC Ingress Connection using the `arn`. For example:

```console
% terraform import aws_apprunner_vpc_ingress_connection.example "arn:aws:apprunner:us-west-2:837424938642:vpcingressconnection/example/b379f86381d74825832c2e82080342fa"
```
