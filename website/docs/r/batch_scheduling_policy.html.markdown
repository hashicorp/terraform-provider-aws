---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_scheduling_policy"
description: |-
  Provides a Batch Scheduling Policy resource.
---

# Resource: aws_batch_scheduling_policy

Provides a Batch Scheduling Policy resource.

## Example Usage

```terraform
resource "aws_batch_scheduling_policy" "example" {
  name = "example"

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A1*"
      weight_factor    = 0.1
    }

    share_distribution {
      share_identifier = "A2"
      weight_factor    = 0.2
    }
  }

  tags = {
    "Name" = "Example Batch Scheduling Policy"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `fairshare_policy` - (Optional) A fairshare policy block specifies the `compute_reservation`, `share_delay_seconds`, and `share_distribution` of the scheduling policy. The `fairshare_policy` block is documented below.
* `name` - (Required) Specifies the name of the scheduling policy.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

A `fairshare_policy` block supports the following arguments:

* `compute_reservation` - (Optional) A value used to reserve some of the available maximum vCPU for fair share identifiers that have not yet been used. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html).
* `share_delay_seconds` - (Optional) The time period to use to calculate a fair share percentage for each fair share identifier in use, in seconds. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html).
* `share_distribution` - (Optional) One or more share distribution blocks which define the weights for the fair share identifiers for the fair share policy. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html). The `share_distribution` block is documented below.

A `share_distribution` block supports the following arguments:

* `share_identifier` - (Required) A fair share identifier or fair share identifier prefix. For more information, see [ShareAttributes](https://docs.aws.amazon.com/batch/latest/APIReference/API_ShareAttributes.html).
* `weight_factor` - (Optional) The weight factor for the fair share identifier. For more information, see [ShareAttributes](https://docs.aws.amazon.com/batch/latest/APIReference/API_ShareAttributes.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name of the scheduling policy.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Batch Scheduling Policy using the `arn`. For example:

```terraform
import {
  to = aws_batch_scheduling_policy.test_policy
  id = "arn:aws:batch:us-east-1:123456789012:scheduling-policy/sample"
}
```

Using `terraform import`, import Batch Scheduling Policy using the `arn`. For example:

```console
% terraform import aws_batch_scheduling_policy.test_policy arn:aws:batch:us-east-1:123456789012:scheduling-policy/sample
```
