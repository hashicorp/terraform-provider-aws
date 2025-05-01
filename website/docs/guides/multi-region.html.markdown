---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Multi-Region Support"
description: |-
  Multi-Region support with the Terraform AWS Provider.
---

# Multi-Region Support

Most AWS resources are Regional – they are created and exist in a single AWS Region, and to manage these resources the Terraform AWS Provider directs API calls to endpoints in the Region. The AWS Region used to provision a resource with the provider is defined in the [provider configuration](https://developer.hashicorp.com/terraform/language/providers/configuration) used by the resource, either implicitly via [environment variables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#environment-variables) or [shared configuration files](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#shared-configuration-and-credentials-files), or explicitly via the [`region` argument](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
Prior to version 6.0.0 of the Terraform AWS Provider in order to manage resources in multiple Regions with a single set of Terraform modules, resources have to use the [`provider` meta-argument](https://developer.hashicorp.com/terraform/language/meta-arguments/resource-provider) along with a separate provider configuration for each Region. For large configurations this adds considerable complexity – today AWS operates in [36 Regions](https://aws.amazon.com/about-aws/global-infrastructure/), with 4 further Regions announced.

To address this, version 6.0.0 of the Terraform AWS Provider adds an additional top-level `region` argument in the [schema](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas) of most Regional resources, data sources, and ephemeral resources, which allows that resource to be managed in a Region other than the one defined in the provider configuration. For those resources that had a pre-existing top-level `region` argument, that argument is now deprecated and in a future version of the provider the `region` argument will be used to implement enhanced multi-Region support. Each such deprecation is noted in a separate section below.

The new top-level `region` argument is [_Optional_ and _Computed_](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string#configurability), with a default value of the Region from the provider configuration. The value of the `region` argument is validated as being in the configured [partition](https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/partitions.html). A change to the argument's value forces resource replacement. To [import](https://developer.hashicorp.com/terraform/cli/import) a resource in a specific Region append `@<region>` to the [import ID](https://developer.hashicorp.com/terraform/language/import#import-id), for example `terraform import aws_vpc.test_vpc vpc-a01106c2@eu-west-1`.

For example, to use a single provider configuration to create S3 buckets in multiple Regions:

```terraform
locals {
  regions = [
    "us-east-1",
    "us-west-2",
    "ap-northeast-1",
    "eu-central-1",
  ]
}

resource "aws_s3_bucket" "example" {
  count  = length(local.regions)
  region = local.regions[count.index]

  bucket = "yournamehere-${local.regions[count.index]}"
}
```

### KMS Replica Key

#### Terraform AWS Provider v5 (and below)

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

#### Terraform AWS Provider v6 (and above)

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


### S3 Bucket Replication Configuration

#### Terraform AWS Provider v5 (and below)

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

#### Terraform AWS Provider v6 (and above)

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

### Global Resources

Some resources (for example [IAM or STS](https://docs.aws.amazon.com/IAM/latest/UserGuide/programming.html#IAMEndpoints) resources) are [global](https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/global-services.html), they exist in all of a partition’s Regions. For such resources no `region` attribute has been added.

#### Global Services

All resources for the following services will be considered _global_:

* Account Management (`aws_account_*`)
* CloudFront (`aws_cloudfront_*` and `aws_cloudfrontkeyvaluestore_*`)
* Global Accelerator (`aws_globalaccelerator_*`)
* IAM (`aws_iam_*`, `aws_rolesanywhere_*` and `aws_caller_identity`)
* Network Manager (`aws_networkmanager_*`)
* Organizations (`aws_organizations_*`)
* Route 53 (`aws_route53_*` and `aws_route53domains_*`)
* Route 53 ARC (`aws_route53recoverycontrolconfig_*` and `aws_route53recoveryreadiness_*`)
* Shield Advanced (`aws_shield_*`)
* WAF Classic (`aws_waf_*`)

#### Global Resources In Regional Services

Some Regional services have subsets of resources that are global:

* Audit Manager
    * `aws_auditmanager_organization_admin_account_registration` is global
* Backup
    * `aws_backup_global_settings` is global
* Billing
    * `aws_billing_service_account` data source is global
* Chime SDK Voice
    * `aws_chimesdkvoice_global_settings` is global
* CloudTrail
    * `aws_cloudtrail_organization_delegated_admin_account` is global
* Detective
    * `aws_detective_organization_admin_account` is global
* Direct Connect
    * `aws_dx_gateway` is global
* EC2
    * `aws_ec2_image_block_public_access` is global
* Firewall Manager
    * `aws_fms_admin_account` is global
* Guard Duty
    * `aws_guardduty_organization_admin_account` is global
* Inspector
    * `aws_inspector2_delegated_admin_account` is global
* IPAM
    * `aws_vpc_ipam_organization_admin_account` is global
* Macie
    * `aws_macie2_organization_admin_account` is global
    * `aws_macie2_organization_configuration` is global
* S3
    * `aws_s3_account_public_access_block` is global
* Security Hub
    * `aws_securityhub_organization_admin_account` is global

#### Meta Data Sources

The `aws_default_tags`, `aws_partition`, and `aws_regions` data sources are effectively global.

The `region` attribute of the `aws_arn` data source stays as-is.

#### Policy Document Data Sources

Some data sources convert HCL into JSON policy documents and are effectively global:

* `aws_cloudwatch_log_data_protection_policy_document`
* `aws_ecr_lifecycle_policy_document`
