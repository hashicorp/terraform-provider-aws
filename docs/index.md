---
layout: "aws"
page_title: "Provider: DATAFYAWS"
description: |-
  Use the Datafy AWS provider to EBS volumes supported by DATAFY. You must configure the provider with the proper credentials before you can use it.
---

# DATAFYAWS Provider

## Example Usage

Terraform 0.13 and later:

```terraform
terraform {
  required_providers {
    datafyaws = {
      source  = "datafy-io/datafyaws"
      version = "~> 5.0"
    }
  }
}

# Configure the AWS Provider
provider "datafyaws" {
  region = "us-east-1"
}

# Create a EBS Volume
resource "aws_ebs_volume" "example" {
  provider = datafyaws

  availability_zone = "us-east-1a"
  size              = 40
}
```

Terraform 0.12 and earlier:

```terraform
# Configure the AWS Provider
provider "datafyaws" {
  version = "~> 5.0"
  region  = "us-east-1"
}

# Create a EBS Volume
resource "aws_ebs_volume" "example" {
  provider = datafyaws

  availability_zone = "us-east-1a"
  size              = 40
}
```

## Configuration

All configuration of DATAFYAWS provider are the same as [AWS provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication-and-configuration)  
The only mandatory addition is _datafy_token_ (that can also be set via environment variables)

## Supported Resources by Datafy
* `aws_ebs_volume`
* `aws_volume_attachment`
* `aws_instance`

