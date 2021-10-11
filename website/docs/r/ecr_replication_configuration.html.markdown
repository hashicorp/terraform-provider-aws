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
        filter      = "a-prefix"
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

* `rule` - (Required) The replication rules for a replication configuration. See [Rule](#rule).

### Rule

* `destination` - (Required) the details of a replication destination. See [Destination](#destination).
* `repository_filter` - (Optional) the details of a replication repository filter. See [Repository Filter](#repository-filter).

### Destination

* `region` - (Required) A Region to replicate to.
* `registry_id` - (Required) The account ID of the destination registry to replicate to.

### Repository Filter

* `filter` - (Required) The repository name prefixe to configure the replication for.
* `registry_id` - (Required) The repository filter type, only support value is `PREFIX_MATCH`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `registry_id` - The registry ID where the replication configuration was created.

## Import

ECR Replication Configuration can be imported using the `registry_id`, e.g.,

```
$ terraform import aws_ecr_replication_configuration.service 012345678912
```
