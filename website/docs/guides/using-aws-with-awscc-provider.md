---
subcategory: ""
layout: "aws"
page_title: "Using Terraform awscc provider with aws provider"
description: |-
  Managing resource tags with the Terraform AWS Provider.
---

# Using AWS & AWSCC Provider Together

~> **NOTE**: Please note the AWSCC provider is currently in technical preview. This means some aspects of its design and implementation are not yet considered stable. We are actively looking for community feedback in order to solidify its form.

The [HashiCorp Terraform AWS Cloud Control Provider](https://registry.terraform.io/providers/hashicorp/awscc/latest) aims to bring Amazon Web Services (AWS) resources to Terraform users faster. The new provider is automatically generated, which means new features and services on AWS can be supported right away. The AWS Cloud Control provider supports hundreds of AWS resources, with more support being added as AWS service teams adopt the Cloud Control API standard.

For Terraform users managing infrastructure on AWS, we expect the AWSCC provider will be used alongside the existing AWS provider. This guide is provided to show guidance and an example of using the providers together to deploy an AWS Cloud WAN Core Network.

For more information about the AWSCC provider, please see the provider documentation in [Terraform Registry](https://registry.terraform.io/providers/hashicorp/awscc/latest)

<!-- TOC depthFrom:2 -->

- [Getting Started with Resource Tags](#getting-started-with-resource-tags)
- [Ignoring Changes to Specific Tags](#ignoring-changes-to-specific-tags)
    - [Ignoring Changes in Individual Resources](#ignoring-changes-in-individual-resources)
    - [Ignoring Changes in All Resources](#ignoring-changes-in-all-resources)
- [Managing Individual Resource Tags](#managing-individual-resource-tags)
- [Propagating Tags to All Resources](#propagating-tags-to-all-resources)

<!-- /TOC -->

## AWS Cloud Wan

In this guide we will deploy [AWS Cloud WAN](https://aws.amazon.com/cloud-wan/) to demonstrate how both AWS & AWSCC can work togther. Cloud WAN is a wide area networking (WAN) service that helps you build, manage, and monitor a unified global network that manages traffic running between resources in your cloud and on-premises environments.

With Cloud WAN, you define network policies that are used to create a global network that spans multiple locations and networks—eliminating the need to configure and manage different networks individually using different technologies. Your network policies can be used to specify which of your Amazon Virtual Private Clouds (VPCs) and on-premises locations you wish to connect through AWS VPN or third-party software-defined WAN (SD-WAN) products, and the Cloud WAN central dashboard generates a complete view of the network to monitor network health, security, and performance. Cloud WAN automatically creates a global network across AWS Regions using Border Gateway Protocol (BGP) so you can easily exchange routes around the world.

For more information on AWS Cloud WAN see <>

## Using Multiple Providers

Terraform can use many providers by specifying them in your `terraform` configuration block:

```terraform
terraform {
  required_version = ">= 0.15.3"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.9.0"
    }
    awscc = {
      source  = "hashicorp/awscc"
      version = ">= 0.17.0"
    }
  }
}
```

The code snippet above informs terraform to download 2 providers as plugins for the current root module, the AWS and AWSCC provider. You can tell which provider is being use by looking at the resource or data source name-prefix. Resources that start with `aws_` use the AWS provider, resources that start with `awscc_` are using the AWSCC provider.

Lets start by building our [global network and core network](https://docs.aws.amazon.com/vpc/latest/cloudwan/cloudwan-networks-working-with.html):

```terraform
resource "aws_networkmanager_global_network" "main" {
  description = "ACME Global Network"

  tags = {
    terraform = "true"
  }
}

resource "awscc_networkmanager_core_network" "main" {
  description = "ACME Core Network"
  global_network_id = aws_networkmanager_global_network.main.id

  tags = [{
    key   = terraform
    value = "true"
  }]
}
```

The global network

The tags for the resource are wholly managed by Terraform except tag keys beginning with `aws:` as these are managed by AWS services and cannot typically be edited or deleted. Any non-AWS tags added to the VPC outside of Terraform will be proposed for removal on the next Terraform execution. Missing tags or those with incorrect values from the Terraform configuration will be proposed for addition or update on the next Terraform execution. Advanced patterns that can adjust these behaviors for special use cases, such as Terraform AWS Provider configurations that affect all resources and the ability to manage resource tags for resources not managed by Terraform, can be found later in this guide.

For most environments and use cases, this is the typical implementation pattern, whether it be in a standalone Terraform configuration or within a [Terraform Module](https://www.terraform.io/docs/modules/). The Terraform configuration language also enables less repetitive configurations via [variables](https://www.terraform.io/docs/configuration/variables.html), [locals](https://www.terraform.io/docs/configuration/locals.html), or potentially a combination of these, e.g.,

```terraform
# Terraform 0.12 and later syntax
variable "additional_tags" {
  default     = {}
  description = "Additional resource tags"
  type        = map(string)
}

resource "aws_vpc" "example" {
  # ... other configuration ...

  # This configuration combines some "default" tags with optionally provided additional tags
  tags = merge(
    var.additional_tags,
    {
      Name = "MyVPC"
    },
  )
}
```

## Ignoring Changes to Specific Tags

Systems outside of Terraform may automatically interact with the tagging associated with AWS resources. These external systems may be for administrative purposes, such as a Configuration Management Database, or the tagging may be required functionality for those systems, such as Kubernetes. This section shows methods to prevent Terraform from showing differences for specific tags.

### Ignoring Changes in Individual Resources

All Terraform resources support the [`lifecycle` configuration block `ignore_changes` argument](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes), which can be used to explicitly ignore all tags changes on a resource beyond an initial configuration or individual tag values.

In this example, the `Name` tag will be added to the VPC on resource creation, however any external changes to the `Name` tag value or the addition/removal of any tag (including the `Name` tag) will be ignored:

```terraform
# Terraform 0.12 and later syntax
resource "aws_vpc" "example" {
  # ... other configuration ...

  tags = {
    Name = "MyVPC"
  }

  lifecycle {
    ignore_changes = [tags]
  }
}
```

In this example, the `Name` and `Owner` tags will be added to the VPC on resource creation, however any external changes to the value of the `Name` tag will be ignored while any changes to other tags (including the `Owner` tag and any additions) will still be proposed:

```terraform
# Terraform 0.12 and later syntax
resource "aws_vpc" "example" {
  # ... other configuration ...

  tags = {
    Name  = "MyVPC"
    Owner = "Operations"
  }

  lifecycle {
    ignore_changes = [tags.Name]
  }
}
```

### Ignoring Changes in All Resources

As of version 2.60.0 of the Terraform AWS Provider, there is support for ignoring tag changes across all resources under a provider. This simplifies situations where certain tags may be externally applied more globally and enhances functionality beyond `ignore_changes` to support cases such as tag key prefixes.

In this example, all resources will ignore any addition of the `LastScanned` tag:

```terraform
provider "aws" {
  # ... potentially other configuration ...

  ignore_tags {
    keys = ["LastScanned"]
  }
}
```

In this example, all resources will ignore any addition of tags with the `kubernetes.io/` prefix, such as `kubernetes.io/cluster/name` or `kubernetes.io/role/elb`:

```terraform
provider "aws" {
  # ... potentially other configuration ...

  ignore_tags {
    key_prefixes = ["kubernetes.io/"]
  }
}
```

Any of the `ignore_tags` configurations can be combined as needed.

The provider ignore tags configuration applies to all Terraform AWS Provider resources under that particular instance (the `default` provider instance in the above cases). If multiple, different Terraform AWS Provider configurations are being used (e.g., [multiple provider instances](https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-instances)), the ignore tags configuration must be added to all applicable provider configurations.

## Managing Individual Resource Tags

Certain Terraform AWS Provider services support a special resource for managing an individual tag on a resource without managing the resource itself. One example is the [`aws_ec2_tag` resource](/docs/providers/aws/r/ec2_tag.html). These resources enable tagging where resources are created outside Terraform such as EC2 Images (AMIs), shared across accounts via Resource Access Manager (RAM), or implicitly created by other means such as EC2 VPN Connections implicitly creating a taggable EC2 Transit Gateway VPN Attachment.

~> **NOTE:** This is an advanced use case and can cause conflicting management issues when improperly implemented. These individual tag resources should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_vpc` and `aws_ec2_tag` to manage tags of the same VPC will cause a perpetual difference where the `aws_vpc` resource will try to remove the tag being added by the `aws_ec2_tag` resource.

-> Not all services supported by the Terraform AWS Provider implement these resources. Browse the Terraform AWS Provider resource documentation pages for a resource with a type ending in `_tag`. If there is a use case where this type of resource is missing, a [feature request](https://github.com/hashicorp/terraform-provider-aws/issues/new?labels=enhancement&template=Feature_Request.md) can be submitted.

```terraform
# Terraform 0.12 and later syntax
# ... other configuration ...

resource "aws_ec2_tag" "example" {
  resource_id = aws_vpn_connection.example.transit_gateway_attachment_id
  key         = "Owner"
  value       = "Operations"
}
```

To manage multiple tags for a resource in this scenario, [`for_each`](https://www.terraform.io/docs/configuration/meta-arguments/for_each.html) can be used:

```terraform
# Terraform 0.12 and later syntax
# ... other configuration ...

resource "aws_ec2_tag" "example" {
  for_each = { "Name" : "MyAttachment", "Owner" : "Operations" }

  resource_id = aws_vpn_connection.example.transit_gateway_attachment_id
  key         = each.key
  value       = each.value
}
```

The inline map provided to `for_each` in the example above is used for brevity, but other Terraform configuration language features similar to those noted at the beginning of this guide can be used to make the example more extensible.

### Propagating Tags to All Resources

As of version 3.38.0 of the Terraform AWS Provider, the Terraform Configuration language also enables provider-level tagging as an alternative to the methods described in the [Getting Started with Resource Tags](#getting-started-with-resource-tags) section above.
This functionality is available for all Terraform AWS Provider resources that currently support `tags`, with the exception of the [`aws_autoscaling_group`](/docs/providers/aws/r/autoscaling_group.html.markdown) resource. Refactoring the use of [variables](https://www.terraform.io/docs/configuration/variables.html) or [locals](https://www.terraform.io/docs/configuration/locals.html) may look like:

```terraform
# Terraform 0.12 and later syntax
provider "aws" {
  # ... other configuration ...
  default_tags {
    tags = {
      Environment = "Production"
      Owner       = "Ops"
    }
  }
}

resource "aws_vpc" "example" {
  # ... other configuration ...

  # This configuration by default will internally combine tags defined
  # within the provider configuration block and those defined here
  tags = {
    Name = "MyVPC"
  }
}
```

In this example, the `Environment` and `Owner` tags defined within the provider configuration block will be added to the VPC on resource creation, in addition to the `Name` tag defined within the VPC resource configuration.
To access all the tags applied to the VPC resource, use the read-only attribute `tags_all`, e.g., `aws_vpc.example.tags_all`.
