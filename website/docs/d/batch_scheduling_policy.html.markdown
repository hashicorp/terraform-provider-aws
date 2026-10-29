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

* `arn` - (Required) ARN of the scheduling policy.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `fair_share_policy` - Fair share policy block of the scheduling policy. The `fair_share_policy` block is documented below.
* `name` - Name of the scheduling policy.
* `tags` - Key-value map of resource tags

### `fair_share_policy` Block

The `fair_share_policy` block exports the following attributes:

* `compute_reservation` - Value used to reserve some of the available maximum vCPU for fair share identifiers that have not yet been used. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html).
* `share_decay_seconds` - Time period to use to calculate a fair share percentage for each fair share identifier in use, in seconds. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html).
* `share_distribution` - One or more share distribution blocks which define the weights for the fair share identifiers for the fair share policy. For more information, see [FairsharePolicy](https://docs.aws.amazon.com/batch/latest/APIReference/API_FairsharePolicy.html). The `share_distribution` block is documented below.

### `share_distribution` Block

The `share_distribution` block exports the following attributes:

* `share_identifier` - Fair share identifier or fair share identifier prefix. For more information, see [ShareAttributes](https://docs.aws.amazon.com/batch/latest/APIReference/API_ShareAttributes.html).
* `weight_factor` - Weight factor for the fair share identifier. For more information, see [ShareAttributes](https://docs.aws.amazon.com/batch/latest/APIReference/API_ShareAttributes.html).
