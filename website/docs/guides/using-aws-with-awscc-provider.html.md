---
subcategory: ""
layout: "aws"
page_title: "Using the Terraform awscc provider with aws provider"
description: |-
  Managing resource tags with the Terraform AWS Provider.
---

# Using AWS & AWSCC Provider Together

~> **NOTE**: The `awscc` provider is currently in technical preview. This means some aspects of its design and implementation are not yet considered stable for production use. We are actively looking for community feedback in order to identify needed improvements.

The [HashiCorp Terraform AWS Cloud Control Provider](https://registry.terraform.io/providers/hashicorp/awscc/latest) aims to bring Amazon Web Services (AWS) resources to Terraform users faster. The new provider is automatically generated, which means new features and services on AWS can be supported right away. The AWS Cloud Control provider supports hundreds of AWS resources, with more support being added as AWS service teams adopt the Cloud Control API standard.

For Terraform users managing infrastructure on AWS, we expect the AWSCC provider will be used alongside the existing AWS provider. This guide is provided to show guidance and an example of using the providers together to deploy an AWS Cloud WAN Core Network.

For more information about the AWSCC provider, please see the provider documentation in [Terraform Registry](https://registry.terraform.io/providers/hashicorp/awscc/latest)

<!-- TOC depthFrom:2 -->

- [AWS CloudWAN Overview](#aws-cloud-wan)
- [Specifying Multiple Providers](#specifying-multiple-providers)
    - [First Look at AWSCC Resources](#first-look-at-awscc-resources)
    - [Using AWS and AWSCC Providers Together](#using-aws-and-awscc-providers-together)

<!-- /TOC -->

## AWS Cloud Wan

In this guide we will deploy [AWS Cloud WAN](https://aws.amazon.com/cloud-wan/) to demonstrate how both AWS & AWSCC can work togther. Cloud WAN is a wide area networking (WAN) service that helps you build, manage, and monitor a unified global network that manages traffic running between resources in your cloud and on-premises environments.

With Cloud WAN, you define network policies that are used to create a global network that spans multiple locations and networks—eliminating the need to configure and manage different networks individually using different technologies. Your network policies can be used to specify which of your Amazon Virtual Private Clouds (VPCs) and on-premises locations you wish to connect through AWS VPN or third-party software-defined WAN (SD-WAN) products, and the Cloud WAN central dashboard generates a complete view of the network to monitor network health, security, and performance. Cloud WAN automatically creates a global network across AWS Regions using Border Gateway Protocol (BGP), so you can easily exchange routes around the world.

For more information on AWS Cloud WAN see [the documentation.](https://docs.aws.amazon.com/vpc/latest/cloudwan/what-is-cloudwan.html)

## Specifying Multiple Providers

Terraform can use many providers at once, as long as they are specified in your `terraform` configuration block:

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
      version = ">= 0.21.0"
    }
  }
}
```

The code snippet above informs terraform to download 2 providers as plugins for the current root module, the AWS and AWSCC provider. You can tell which provider is being use by looking at the resource or data source name-prefix. Resources that start with `aws_` use the AWS provider, resources that start with `awscc_` are using the AWSCC provider.

### First look at AWSCC resources

Lets start by building our [global network](https://aws.amazon.com/about-aws/global-infrastructure/global_network/) which will house our core network.

```terraform
locals {
  terraform_tag = [{
    key   = "terraform"
    value = "true"
  }]
}

resource "awscc_networkmanager_global_network" "main" {
  description = "My Global Network"
  tags = concat(local.terraform_tag,
    [{
      key   = "Name"
      value = "My Global Network"
    }]
  )
}
```

Above, we define a `awscc_networkmanager_global_network` with 2 tags and a description. AWSCC resources use the [standard AWS tag format](https://docs.aws.amazon.com/general/latest/gr/aws_tagging.html) which is expressed in HCL as a list of maps with 2 keys. We want to reuse the `terraform = true` tag so we define it as a `local` then we use [concat](https://www.terraform.io/language/functions/concat) to join the list of tags together.

### Using AWS and AWSCC providers together

Next we will create a [core network](https://docs.aws.amazon.com/vpc/latest/cloudwan/cloudwan-core-network-policy.html) using an AWSCC resource `awscc_networkmanager_core_network` and an AWS data source `data.aws_networkmanager_core_network_policy_document` which allows users to write HCL to generate the json policy used as the [core policy network](https://docs.aws.amazon.com/vpc/latest/cloudwan/cloudwan-policies-json.html).

```
resource "awscc_networkmanager_core_network" "main" {
  description       = "My Core Network"
  global_network_id = awscc_networkmanager_global_network.main.id
  policy_document   = data.aws_networkmanager_core_network_policy_document.main.json
  tags              = local.terraform_tag
}

data "aws_networkmanager_core_network_policy_document" "main" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = "us-east-1"
      asn      = 64512
    }
  }

  segments {
    name                          = "shared"
    description                   = "Segment for shared services"
    require_attachment_acceptance = true
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "or"

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "segment"
      value    = "shared"
    }
    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}
```

Thanks to Terraform's plugin design, the providers work together seemlessly!
