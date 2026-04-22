---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Enhanced Region Support"
description: |-
  Enhanced Region support with the Terraform AWS Provider.
---

# Enhanced Region Support

Version 6.0.0 of the Terraform AWS Provider adds `region` to most resources making it significantly easier to manage infrastructure across AWS Regions without requiring multiple provider configurations.

<!-- TOC depthFrom:2 depthTo:2 -->

- [What's new](#whats-new)
- [What's not changing](#whats-not-changing)
- [Can I use `region` in every resource?](#can-i-use-region-in-every-resource)
- [Why make this change](#why-make-this-change)
- [How `region` works](#how-region-works)
- [Migrating from multiple provider configurations](#migrating-from-multiple-provider-configurations)
- [Before and after examples using `region`](#before-and-after-examples-using-region)
- [Non-region-aware resources](#non-region-aware-resources)

<!-- /TOC -->

## What's new

As of v6.0.0, most existing resources, data sources, and ephemeral resources are now Region-aware, meaning they support a new top-level `region` argument. This allows you to manage a resource in a Region different from the one specified in the provider configuration without requiring multiple provider blocks. See [How `region` works](#how-region-works) for details.

For example, if your provider is configured for `us-east-1`, you can now manage a VPC in `us-west-2` without defining an additional provider block:

```terraform
resource "aws_vpc" "peer" {
  region     = "us-west-2"
  cidr_block = "10.1.0.0/16"
}
```

## What's _not_ changing

_Pre-v6.0.0 configurations that use provider blocks per Region remain valid in v6.0.0 and are not deprecated._

You can still define the Region at the provider level using any of the existing methods—for example, through the AWS [config file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html), [provider configuration](https://developer.hashicorp.com/terraform/language/providers/configuration), [environment variables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#environment-variables), [shared configuration files](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#shared-configuration-and-credentials-files), or explicitly using the `provider`’s [`region`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).

## Can I use `region` in every resource?

No. While most resources are now Region-aware, there are exceptions. These include a few resources that already had a `region` and resources that are inherently global. See [Non-region-aware resources](#non-region-aware-resources).

## Why make this change

Before version 6.0.0, managing infrastructure across multiple Regions required a separate provider configuration for each Region. This approach led to complex and repetitive configurations, especially for large infrastructures—AWS currently operates in [36 Regions](https://aws.amazon.com/about-aws/global-infrastructure/), with more announced. Additionally, each provider configuration adds overhead in terms of memory and compute resources.

See the [examples](#before-and-after-examples-using-region) below for a comparison of configurations before and after introducing `region`.

## How `region` works

The new top-level `region` argument is [_Optional_ and _Computed_](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string#configurability), and defaults to the Region specified in the provider configuration. Its value is validated to ensure it belongs to the configured [partition](https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/partitions.html).

**Changing the value of `region` will force resource replacement.**

**Removing `region` does not force resource replacement. The prior value of `region` stored in Terraform state will be used.**

To [import](https://developer.hashicorp.com/terraform/cli/import) a resource in a specific Region, append `@<region>` to the [import ID](https://developer.hashicorp.com/terraform/language/import#import-id)—for example:

```sh
terraform import aws_vpc.test_vpc vpc-a01106c2@eu-west-1
```

## Migrating from multiple provider configurations

To migrate from a separate provider configuration for each Region to a single provider configuration block and per-resource `region` values you must ensure that Terraform state is refreshed before editing resource configuration:

1. Upgrade to v6.0.0
2. Run a Terraform apply in [refresh-only mode](https://developer.hashicorp.com/terraform/cli/commands/plan#planning-modes) -- `terraform apply -refresh-only`
3. Modify the affected resource configurations, replacing the [`provider` meta-argument](https://developer.hashicorp.com/terraform/language/meta-arguments/resource-provider) with a `region` argument

## Before and after examples using `region`

### Cross-region VPC peering

<details>
<summary>Before, Pre-v6.0.0</summary>
<p>

```terraform
provider "aws" {
  region = "us-east-1"
}

provider "aws" {
  alias  = "peer"
  region = "us-west-2"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc" "peer" {
  provider   = aws.peer
  cidr_block = "10.1.0.0/16"
}

data "aws_caller_identity" "peer" {
  provider = aws.peer
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "peer" {
  vpc_id        = aws_vpc.main.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  peer_region   = "us-west-2"
  auto_accept   = false
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider                  = aws.peer
  vpc_peering_connection_id = aws_vpc_peering_connection.peer.id
  auto_accept               = true
}
```

</p>
</details>

<details>
<summary>After, v6.0.0+</summary>
<p>

```terraform
provider "aws" {
  region = "us-east-1"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc" "peer" {
  region     = "us-west-2"
  cidr_block = "10.1.0.0/16"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "peer" {
  vpc_id      = aws_vpc.main.id
  peer_vpc_id = aws_vpc.peer.id
  peer_region = "us-west-2"
  auto_accept = false
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  region                    = "us-west-2"
  vpc_peering_connection_id = aws_vpc_peering_connection.peer.id
  auto_accept               = true
}
```

</p>
</details>

### KMS replica key

<details>
<summary>Before, Pre-v6.0.0</summary>
<p>

```terraform
provider "aws" {
  alias  = "primary"
  region = "us-east-1"
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_kms_key" "primary" {
  provider = aws.primary

  description             = "Multi-Region primary key"
  deletion_window_in_days = 30
  multi_region            = true
}

resource "aws_kms_replica_key" "replica" {
  description             = "Multi-Region replica key"
  deletion_window_in_days = 7
  primary_key_arn         = aws_kms_key.primary.arn
}
```

</p>
</details>

<details>
<summary>After, v6.0.0</summary>
<p>

```terraform
provider "aws" {
  region = "us-west-2"
}

resource "aws_kms_key" "primary" {
  region = "us-east-1"

  description             = "Multi-Region primary key"
  deletion_window_in_days = 30
  multi_region            = true
}

resource "aws_kms_replica_key" "replica" {
  description             = "Multi-Region replica key"
  deletion_window_in_days = 7
  primary_key_arn         = aws_kms_key.primary.arn
}
```

</p>
</details>

### S3 bucket replication configuration

<details>
<summary>Before, Pre-v6.0.0</summary>
<p>

```terraform
provider "aws" {
  region = "eu-west-1"
}

provider "aws" {
  alias  = "central"
  region = "eu-central-1"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "replication" {
  name               = "tf-iam-role-replication-12345"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "replication" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetReplicationConfiguration",
      "s3:ListBucket",
    ]

    resources = [aws_s3_bucket.source.arn]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObjectVersionForReplication",
      "s3:GetObjectVersionAcl",
      "s3:GetObjectVersionTagging",
    ]

    resources = ["${aws_s3_bucket.source.arn}/*"]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:ReplicateObject",
      "s3:ReplicateDelete",
      "s3:ReplicateTags",
    ]

    resources = ["${aws_s3_bucket.destination.arn}/*"]
  }
}

resource "aws_iam_policy" "replication" {
  name   = "tf-iam-role-policy-replication-12345"
  policy = data.aws_iam_policy_document.replication.json
}

resource "aws_iam_role_policy_attachment" "replication" {
  role       = aws_iam_role.replication.name
  policy_arn = aws_iam_policy.replication.arn
}

resource "aws_s3_bucket" "destination" {
  bucket = "tf-test-bucket-destination-12345"
}

resource "aws_s3_bucket_versioning" "destination" {
  bucket = aws_s3_bucket.destination.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "source" {
  provider = aws.central
  bucket   = "tf-test-bucket-source-12345"
}

resource "aws_s3_bucket_acl" "source_bucket_acl" {
  provider = aws.central

  bucket = aws_s3_bucket.source.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "source" {
  provider = aws.central

  bucket = aws_s3_bucket.source.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "replication" {
  provider = aws.central
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.source]

  role   = aws_iam_role.replication.arn
  bucket = aws_s3_bucket.source.id

  rule {
    id = "examplerule"

    filter {
      prefix = "example"
    }

    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }
}
```

</p>
</details>

<details>
<summary>After, v6.0.0</summary>
<p>

```terraform
provider "aws" {
  region = "eu-west-1"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "replication" {
  name               = "tf-iam-role-replication-12345"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "replication" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetReplicationConfiguration",
      "s3:ListBucket",
    ]

    resources = [aws_s3_bucket.source.arn]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObjectVersionForReplication",
      "s3:GetObjectVersionAcl",
      "s3:GetObjectVersionTagging",
    ]

    resources = ["${aws_s3_bucket.source.arn}/*"]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:ReplicateObject",
      "s3:ReplicateDelete",
      "s3:ReplicateTags",
    ]

    resources = ["${aws_s3_bucket.destination.arn}/*"]
  }
}

resource "aws_iam_policy" "replication" {
  name   = "tf-iam-role-policy-replication-12345"
  policy = data.aws_iam_policy_document.replication.json
}

resource "aws_iam_role_policy_attachment" "replication" {
  role       = aws_iam_role.replication.name
  policy_arn = aws_iam_policy.replication.arn
}

resource "aws_s3_bucket" "destination" {
  bucket = "tf-test-bucket-destination-12345"
}

resource "aws_s3_bucket_versioning" "destination" {
  bucket = aws_s3_bucket.destination.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "source" {
  region = "eu-central-1"

  bucket = "tf-test-bucket-source-12345"
}

resource "aws_s3_bucket_acl" "source_bucket_acl" {
  region = "eu-central-1"

  bucket = aws_s3_bucket.source.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "source" {
  region = "eu-central-1"

  bucket = aws_s3_bucket.source.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "replication" {
  region = "eu-central-1"

  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.source]

  role   = aws_iam_role.replication.arn
  bucket = aws_s3_bucket.source.id

  rule {
    id = "examplerule"

    filter {
      prefix = "example"
    }

    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }
}
```

</p>
</details>

## Non-region-aware resources

This section lists resources that are not Region-aware, meaning `region` has not been added to them.

Some resources, such as [IAM and STS](https://docs.aws.amazon.com/IAM/latest/UserGuide/programming.html#IAMEndpoints), are [global](https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/global-services.html) and exist in all Regions within a partition.

Other resources are not Region-aware because they already had a top-level `region`, are inherently global, or because adding `region` would not be appropriate for other reasons.

### Resources deprecating `region`

The following regional resources and data sources had a top-level `region` prior to version 6.0.0. It is now deprecated and will be replaced in a future version to support the new Region-aware behavior.

* `aws_cloudformation_stack_set_instance` resource
* `aws_config_aggregate_authorization` resource
* `aws_dx_hosted_connection` resource
* `aws_region` data source
* `aws_s3_bucket` data source
* `aws_servicequotas_template` resource
* `aws_servicequotas_templates` data source
* `aws_ssmincidents_replication_set` resource and data source
* `aws_vpc_endpoint_service` data source
* `aws_vpc_peering_connection` data source

### Global services

All resources for the following services are considered _global_:

* Account Management (`aws_account_*`)
* ARC Region Switch (`aws_arcregionswitch_*`)
* Billing (`aws_billing_*`)
* Billing and Cost Management Data Exports (`aws_bcmdataexports_*`)
* Budgets (`aws_budgets_*`)
* CloudFront (`aws_cloudfront_*` and `aws_cloudfrontkeyvaluestore_*`)
* Cost Explorer (`aws_ce_*`)
* Cost Optimization Hub (`aws_costoptimizationhub_*`)
* Cost and Usage Report (`aws_cur_*`)
* Global Accelerator (`aws_globalaccelerator_*`)
* IAM (`aws_iam_*`, `aws_rolesanywhere_*` and `aws_caller_identity`)
* Invoicing (`aws_invoicing_*`)
* Network Manager (`aws_networkmanager_*`)
* Organizations (`aws_organizations_*`)
* Price List (`aws_pricing_*`)
* Route 53 (`aws_route53_*` and `aws_route53domains_*`)
* Route 53 ARC (`aws_route53recoverycontrolconfig_*` and `aws_route53recoveryreadiness_*`)
* Savings Plans (`aws_savingsplans_*`)
* Shield Advanced (`aws_shield_*`)
* User Notifications (`aws_notifications_*`)
* User Notifications Contacts (`aws_notificationscontacts_*`)
* WAF Classic (`aws_waf_*`)

### Global resources in regional services

Some regional services have a subset of resources that are global:

| Service | Type | Name |
|---|---|---|
| Backup | Resource | `aws_backup_global_settings` |
| Chime SDK Voice | Resource | `aws_chimesdkvoice_global_settings` |
| CloudTrail | Resource | `aws_cloudtrail_organization_delegated_admin_account` |
| Direct Connect | Resource | `aws_dx_gateway` |
| Direct Connect | Data Source | `aws_dx_gateway` |
| Firewall Manager | Resource | `aws_fms_admin_account` |
| IPAM | Resource | `aws_vpc_ipam_organization_admin_account` |
| Resource Access Manager | Resource | `aws_ram_sharing_with_organization` |
| S3 | Data Source | `aws_canonical_user_id` |
| S3 | Resource | `aws_s3_account_public_access_block` |
| S3 | Data Source | `aws_s3_account_public_access_block` |
| Service Catalog | Resource | `aws_servicecatalog_organizations_access` |

### Meta data sources

The `aws_default_tags`, `aws_partition`, and `aws_regions` data sources are effectively global.

`region` of the `aws_arn` data source stays as-is.

### Policy Document Data Sources

Some data sources convert HCL into JSON policy documents and are effectively global:

* `aws_cloudwatch_log_data_protection_policy_document`
* `aws_ecr_lifecycle_policy_document`
