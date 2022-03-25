---
subcategory: "ECR"
layout: "aws"
page_title: "AWS: aws_ecr_replication_configuration"
description: |-
  Provides an Elastic Container Registry Replication Configuration.
---

# Resource: aws_ecr_replication_configuration

Provides an Elastic Container Registry Replication Configuration.

## Example Usage

```terraform
data "aws_caller_identity" "current" {}

data "aws_regions" "example" {}

resource "aws_ecr_replication_configuration" "example" {
  replication_configuration {
    rule {
      destination {
        region      = data.aws_regions.example.names[0]
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
```

## Multiple Region Usage

```terraform
data "aws_caller_identity" "current" {}

data "aws_regions" "example" {}

resource "aws_ecr_replication_configuration" "example" {
  replication_configuration {
    rule {
      destination {
        region      = data.aws_regions.example.names[0]
        registry_id = data.aws_caller_identity.current.account_id
      }

      destination {
        region      = data.aws_regions.example.names[1]
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
```

## Repository Filter Usage

```terraform
data "aws_caller_identity" "current" {}

data "aws_regions" "example" {}

resource "aws_ecr_replication_configuration" "example" {
  replication_configuration {
    rule {
      destination {
        region      = data.aws_regions.example.names[0]
        registry_id = data.aws_caller_identity.current.account_id
      }

      repository_filter {
        filter      = "prod-microservice"
        filter_type = "PREFIX_MATCH"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `replication_configuration` - (Required) Replication configuration for a registry. See [Replication Configuration](#replication-configuration).

### Replication Configuration

* `rule` - (Required) The replication rules for a replication configuration. A maximum of 10 are allowed per `replication_configuration`. See [Rule](#rule)

### Rule

* `destination` - (Required) the details of a replication destination. A maximum of 25 are allowed per `rule`. See [Destination](#destination).
* `repository_filter` - (Optional) filters for a replication rule. See [Repository Filter](#repository-filter).

### Destination

* `region` - (Required) A Region to replicate to.
* `registry_id` - (Required) The account ID of the destination registry to replicate to.

### Repository Filter

* `filter` - (Required) The repository filter details.
* `filter_type` - (Required) The repository filter type. The only supported value is `PREFIX_MATCH`, which is a repository name prefix specified with the filter parameter.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `registry_id` - The registry ID where the replication configuration was created.

## Import

ECR Replication Configuration can be imported using the `registry_id`, e.g.,

```
$ terraform import aws_ecr_replication_configuration.service 012345678912
```
