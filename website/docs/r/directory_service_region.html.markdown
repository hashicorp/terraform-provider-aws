---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_region"
description: |-
    Manages a replicated Region and directory for Multi-Region replication.
---

# Resource: aws_directory_service_region

Manages a replicated Region and directory for Multi-Region replication.
Multi-Region replication is only supported for the Enterprise Edition of AWS Managed Microsoft AD.

## Example Usage

```terraform
provider "aws" {
  region = "us-west-2"
}

provider "aws" {
  alias  = "secondary"
  region = "us-east-2"
}

data "aws_region" "example" {
  provider = aws.secondary
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "Primary"
  }
}

resource "aws_subnet" "example" {
  count = 2

  vpc_id            = aws_vpc.example.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.example.cidr_block, 8, count.index)

  tags = {
    Name = "Primary"
  }
}

resource "aws_directory_service_directory" "example" {
  name     = "example.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.example.id
    subnet_ids = aws_subnet.example[*].id
  }
}

data "aws_availability_zones" "available-secondary" {
  provider = aws.secondary

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "example-secondary" {
  provider = aws.secondary

  cidr_block = "10.1.0.0/16" # Can't overlap with primary's VPC.

  tags = {
    Name = "Secondary"
  }
}

resource "aws_subnet" "example-secondary" {
  provider = aws.secondary

  count = 2

  vpc_id            = aws_vpc.example-secondary.id
  availability_zone = data.aws_availability_zones.available-secondary.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.example-secondary.cidr_block, 8, count.index)

  tags = {
    Name = "Secondary"
  }
}

resource "aws_directory_service_region" "example" {
  directory_id = aws_directory_service_directory.example.id
  region_name  = data.aws_region.example.name

  vpc_settings {
    vpc_id     = aws_vpc.example-secondary.id
    subnet_ids = aws_subnet.example-secondary[*].id
  }

  tags = {
    Name = "Secondary"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `desired_number_of_domain_controllers` - (Optional) The number of domain controllers desired in the replicated directory. Minimum value of `2`.
* `directory_id` - (Required) The identifier of the directory to which you want to add Region replication.
* `region_name` - (Required) The name of the Region where you want to add domain controllers for replication.
* `tags` - (Optional) Map of tags to assign to this resource. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_settings` - (Required) VPC information in the replicated Region. Detailed below.

### `vpc_settings`

* `subnet_ids` - (Required) The identifiers of the subnets for the directory servers.
* `vpc_id` - (Optional) The identifier of the VPC in which to create the directory.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_directory_service_region` provides the following [Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) configuration options:

- `create` - (Default `180 minutes`) Used for Region addition
- `update` - (Default `90 minutes`) Used for replicated directory update
- `delete` - (Default `90 minutes`) Used for Region removal

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Replicated Regions using directory ID,Region name. For example:

```terraform
import {
  to = aws_directory_service_region.example
  id = "d-9267651497,us-east-2"
}
```

Using `terraform import`, import Replicated Regions using directory ID,Region name. For example:

```console
% terraform import aws_directory_service_region.example d-9267651497,us-east-2
```
