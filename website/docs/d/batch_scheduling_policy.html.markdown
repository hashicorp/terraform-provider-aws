---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_scheduling_policy"
description: |-
    Provides details about a Batch Scheduling Policy
---

# Data Source: aws_batch_scheduling_policy

The Batch Scheduling Policy data source allows access to details of a specific Scheduling Policy within AWS Batch.

## Example Usage

```terraform
data "aws_batch_scheduling_policy" "test" {
  arn = "arn:aws:batch:us-east-1:012345678910:scheduling-policy/example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Required) ARN of the scheduling policy.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `fairshare_policy` - Fairshare policy block specifies the `compute_reservation`, `share_delay_seconds`, and `share_distribution` of the scheduling policy. The `fairshare_policy` block is documented below.
* `name` - Name of the scheduling policy.
* `tags` - Key-value map of resource tags

A `fairshare_policy` block supports the following arguments:

* `compute_reservation` - Value used to reserve some of the available maximum vCPU for fair share identifiers that have not yet been used. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html).
* `share_delay_seconds` - Time period to use to calculate a fair share percentage for each fair share identifier in use, in seconds. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html).
* `share_distribution` - One or more share distribution blocks which define the weights for the fair share identifiers for the fair share policy. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html). The `share_distribution` block is documented below.

A `share_distribution` block supports the following arguments:

* `share_identifier` - Fair share identifier or fair share identifier prefix. For more information, see [ShareAttributes](https://docs.aws.amazon.com/batch/latest/APIReference/API_ShareAttributes.html).
* `weight_factor` - Weight factor for the fair share identifier. For more information, see [ShareAttributes](https://docs.aws.amazon.com/batch/latest/APIReference/API_ShareAttributes.html).
