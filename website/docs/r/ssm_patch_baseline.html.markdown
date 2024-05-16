---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_patch_baseline"
description: |-
  Provides an SSM Patch Baseline resource
---

# Resource: aws_ssm_patch_baseline

Provides an SSM Patch Baseline resource.

~> **NOTE on Patch Baselines:** The `approved_patches` and `approval_rule` are
both marked as optional fields, but the Patch Baseline requires that at least one
of them is specified.

## Example Usage

### Basic Usage

Using `approved_patches` only.

```terraform
resource "aws_ssm_patch_baseline" "production" {
  name             = "patch-baseline"
  approved_patches = ["KB123456"]
}
```

### Advanced Usage, specifying patch filters

```terraform
resource "aws_ssm_patch_baseline" "production" {
  name             = "patch-baseline"
  description      = "Patch Baseline Description"
  approved_patches = ["KB123456", "KB456789"]
  rejected_patches = ["KB987654"]

  global_filter {
    key    = "PRODUCT"
    values = ["WindowsServer2008"]
  }

  global_filter {
    key    = "CLASSIFICATION"
    values = ["ServicePacks"]
  }

  global_filter {
    key    = "MSRC_SEVERITY"
    values = ["Low"]
  }

  approval_rule {
    approve_after_days = 7
    compliance_level   = "HIGH"

    patch_filter {
      key    = "PRODUCT"
      values = ["WindowsServer2016"]
    }

    patch_filter {
      key    = "CLASSIFICATION"
      values = ["CriticalUpdates", "SecurityUpdates", "Updates"]
    }

    patch_filter {
      key    = "MSRC_SEVERITY"
      values = ["Critical", "Important", "Moderate"]
    }
  }

  approval_rule {
    approve_after_days = 7

    patch_filter {
      key    = "PRODUCT"
      values = ["WindowsServer2012"]
    }
  }
}
```

### Advanced usage, specifying Microsoft application and Windows patch rules

```terraform
resource "aws_ssm_patch_baseline" "windows_os_apps" {
  name             = "WindowsOSAndMicrosoftApps"
  description      = "Patch both Windows and Microsoft apps"
  operating_system = "WINDOWS"

  approval_rule {
    approve_after_days = 7

    patch_filter {
      key    = "CLASSIFICATION"
      values = ["CriticalUpdates", "SecurityUpdates"]
    }

    patch_filter {
      key    = "MSRC_SEVERITY"
      values = ["Critical", "Important"]
    }
  }

  approval_rule {
    approve_after_days = 7

    patch_filter {
      key    = "PATCH_SET"
      values = ["APPLICATION"]
    }

    # Filter on Microsoft product if necessary
    patch_filter {
      key    = "PRODUCT"
      values = ["Office 2013", "Office 2016"]
    }
  }
}
```

### Advanced usage, specifying alternate patch source repository

```terraform
resource "aws_ssm_patch_baseline" "al_2017_09" {
  name             = "Amazon-Linux-2017.09"
  description      = "My patch repository for Amazon Linux 2017.09"
  operating_system = "AMAZON_LINUX"

  approval_rule {
    # ...
  }

  source {
    name          = "My-AL2017.09"
    products      = ["AmazonLinux2017.09"]
    configuration = <<EOF
[amzn-main]
name=amzn-main-Base
mirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list
mirrorlist_expire=300
metadata_expire=300
priority=10
failovermethod=priority
fastestmirror_enabled=0
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga
enabled=1
retries=3
timeout=5
report_instanceid=yes
EOF
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the patch baseline.

The following arguments are optional:

* `approval_rule` - (Optional) Set of rules used to include patches in the baseline. Up to 10 approval rules can be specified. See [`approval_rule`](#approval_rule-block) below.
* `approved_patches_compliance_level` - (Optional) Compliance level for approved patches. This means that if an approved patch is reported as missing, this is the severity of the compliance violation. Valid values are `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`, `INFORMATIONAL`, `UNSPECIFIED`. The default value is `UNSPECIFIED`.
* `approved_patches_enable_non_security` - (Optional) Whether the list of approved patches includes non-security updates that should be applied to the instances. Applies to Linux instances only.
* `approved_patches` - (Optional) List of explicitly approved patches for the baseline. Cannot be specified with `approval_rule`.
* `description` - (Optional) Description of the patch baseline.
* `global_filter` - (Optional) Set of global filters used to exclude patches from the baseline. Up to 4 global filters can be specified using Key/Value pairs. Valid Keys are `PRODUCT`, `CLASSIFICATION`, `MSRC_SEVERITY`, and `PATCH_ID`.
* `operating_system` - (Optional) Operating system the patch baseline applies to. Valid values are `ALMA_LINUX`, `AMAZON_LINUX`, `AMAZON_LINUX_2`, `AMAZON_LINUX_2022`, `AMAZON_LINUX_2023`, `CENTOS`, `DEBIAN`, `MACOS`, `ORACLE_LINUX`, `RASPBIAN`, `REDHAT_ENTERPRISE_LINUX`, `ROCKY_LINUX`, `SUSE`, `UBUNTU`, and `WINDOWS`. The default value is `WINDOWS`.
* `rejected_patches_action` - (Optional) Action for Patch Manager to take on patches included in the `rejected_patches` list. Valid values are `ALLOW_AS_DEPENDENCY` and `BLOCK`.
* `rejected_patches` - (Optional) List of rejected patches.
* `source` - (Optional) Configuration block with alternate sources for patches. Applies to Linux instances only. See [`source`](#source-block) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `approval_rule` Block

The `approval_rule` block supports:

* `approve_after_days` - (Optional) Number of days after the release date of each patch matched by the rule the patch is marked as approved in the patch baseline. Valid Range: 0 to 100. Conflicts with `approve_until_date`.
* `approve_until_date` - (Optional) Cutoff date for auto approval of released patches. Any patches released on or before this date are installed automatically. Date is formatted as `YYYY-MM-DD`. Conflicts with `approve_after_days`
* `compliance_level` - (Optional) Compliance level for patches approved by this rule. Valid values are `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`, `INFORMATIONAL`, and `UNSPECIFIED`. The default value is `UNSPECIFIED`.
* `enable_non_security` - (Optional) Boolean enabling the application of non-security updates. The default value is `false`. Valid for Linux instances only.
* `patch_filter` - (Required) Patch filter group that defines the criteria for the rule. Up to 5 patch filters can be specified per approval rule using Key/Value pairs. Valid combinations of these Keys and the `operating_system` value can be found in the [SSM DescribePatchProperties API Reference](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_DescribePatchProperties.html). Valid Values are exact values for the patch property given as the key, or a wildcard `*`, which matches all values. `PATCH_SET` defaults to `OS` if unspecified

### `source` Block

The `source` block supports:

* `configuration` - (Required) Value of the yum repo configuration. For information about other options available for your yum repository configuration, see the [`dnf.conf` documentation](https://man7.org/linux/man-pages/man5/dnf.conf.5.html)
* `name` - (Required) Name specified to identify the patch source.
* `products` - (Required) Specific operating system versions a patch repository applies to, such as `"Ubuntu16.04"`, `"AmazonLinux2016.09"`, `"RedhatEnterpriseLinux7.2"` or `"Suse12.7"`. For lists of supported product values, see [PatchFilter](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_PatchFilter.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the baseline.
* `id` - ID of the baseline.
* `json` - JSON definition of the baseline.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM Patch Baselines using their baseline ID. For example:

```terraform
import {
  to = aws_ssm_patch_baseline.example
  id = "pb-12345678"
}
```

Using `terraform import`, import SSM Patch Baselines using their baseline ID. For example:

```console
% terraform import aws_ssm_patch_baseline.example pb-12345678
```
