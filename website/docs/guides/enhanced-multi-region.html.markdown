---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Enhanced Multi-Region Support"
description: |-
  Enhanced multi-Region support with the Terraform AWS Provider.
---

# Enhanced Multi-Region Support

Most AWS resources are Regional – they are created and exist in a single AWS Region, and to manage these resources the Terraform AWS Provider directs API calls to endpoints in the Region. The AWS Region used to provision a resource with the provider is defined in the [provider configuration](https://developer.hashicorp.com/terraform/language/providers/configuration) used by the resource, either implicitly via [environment variables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#environment-variables) or [shared configuration files](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#shared-configuration-and-credentials-files), or explicitly via the [`region` argument](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
To manage resources in multiple Regions with a single set of Terraform modules, resources had to use the [`provider` meta-argument](https://developer.hashicorp.com/terraform/language/meta-arguments/resource-provider) along with a separate provider configuration for each Region. For large configurations this adds considerable complexity – today AWS operates in [36 Regions](https://aws.amazon.com/about-aws/global-infrastructure/), with 4 further Regions announced.

To address this, there is now an additional top-level `region` argument in the [schema](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas) of most Regional resources, data sources, and ephemeral resources, which allows that resource to be managed in a Region other than the one defined in the provider configuration. For those resources that had a pre-existing top-level `region` argument, that argument is now deprecated and in a future version of the provider the `region` argument will be used to implement enhanced multi-Region support. Each such deprecation is noted in a separate section below.

The new top-level `region` argument is [_Optional_ and _Computed_](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string#configurability), with a default value of the Region from the provider configuration. The value of the `region` argument is validated as being in the configured [partition](https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/partitions.html). A change to the argument's value forces resource replacement. To [import](https://developer.hashicorp.com/terraform/cli/import) a resource in a specific Region append `@<region>` to the [import ID](https://developer.hashicorp.com/terraform/language/import#import-id), for example `terraform import aws_vpc.test_vpc vpc-a01106c2@eu-west-1`.

For example, to use a singe provider configuration to create S3 buckets in multiple Regions:

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