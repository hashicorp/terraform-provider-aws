# Frequently Asked Questions

<!-- markdownlint-disable no-trailing-punctuation -->

### Who are the maintainers?

The HashiCorp Terraform AWS provider team is :

* Mary Cutrali, Product Manager - GitHub [@maryelizbeth](https://github.com/maryelizbeth) Twitter [@marycutrali](https://twitter.com/marycutrali)
* Simon Davis, Engineering Manager - GitHub [@breathingdust](https://github.com/breathingdust)
* Justin Retzolk, Technical Community Manager - GitHub [@justinretzolk](https://github.com/justinretzolk)
* Angie Pinilla, Engineer - GitHub [@angie44](https://github.com/angie44)
* Dirk Avery, Engineer - GitHub [@YakDriver](https://github.com/yakdriver)
* Graham Davison, Engineer - GitHub [@gdavison](https://github.com/gdavison)
* Kerim Satirli, Developer Advocate - GitHub [@ksatirli](https://github.com/ksatirli)
* Kit Ewbank, Engineer - GitHub [@ewbankkit](https://github.com/ewbankkit)
* Zoe Helding, Engineer - GitHub [@zhelding](https://github.com/zhelding)
* Adrian Johnson, Engineer - GitHub [@johnsonaj](https://github.com/johnsonaj)

### Why isn’t my PR merged yet?

Unfortunately, due to the volume of issues and new pull requests we receive, we are unable to give each one the full attention that we would like. We always focus on the contributions that provide the greatest value to the most community members. For more information on how we prioritize pull requests, see the [prioritization guide](./prioritization.md).

### How do you decide what gets merged for each release?

We have a large backlog of pull requests to get through and the team are moving through them as quick as we can. All pull requests must be reviewed by a HashiCorp engineer before inclusion. This is to ensure that the design of the addition fits with what provider users have come to expect, and to ensure that testing and best practices are adhered to. This is particularly important for such a large codebase, to ensure that we sustain its maintainability as its grows.

The number one factor we look at when deciding what issues to look at are your 👍 [reactions](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original issue/PR description as these can be [easily discovered](https://github.com/hashicorp/terraform-provider-aws/issues?q=is%3Aopen+sort%3Areactions-%2B1-desc). Comments that further explain desired use cases or poor user experience are also heavily factored. The items with the most support are always on our radar, and we commit to keep the community updated on their status and potential timelines.

We publish a [roadmap](../../ROADMAP.md) every quarter which describes major themes or specific product areas of focus. What is excluded from the public roadmap is work performed under NDA with AWS on new services, and any ad-hoc work we pick up during the quarter. This ad-hoc work can be responding to bugs, gardening day activity, customer prioritization, and technical debt items.

We also are investing time to improve the contributing experience by improving documentation, adding more linter coverage to ensure that incoming PR's can be in as good shape as possible. This will allow us to get through them quicker.

### My PR hasn't been merged and it now has merge conflicts/failed checks, should I keep it up to date?

We realize that sometimes pull requests sit for a considerable amount of time without being addressed. During this time period they may accumulate merge conflicts and failed linter checks as the provider codebase moves forward. As maintainers we have no expectation that you keep your PR up to date, these issues will be addressed at review time most often by the maintainers themselves. Obviously we would hope that your PR is mergeable when first raised! The mergeability of the PR does not affect its prioritization for review.

### How often do you release?

We release weekly on Thursday. We release often to ensure we can bring value to the community at a frequent cadence and to ensure we are in a good place to react to AWS region launches and service announcements.

### Backward Compatibility Promise

Our policy is described on the Terraform website [here](https://www.terraform.io/docs/extend/best-practices/versioning.html). While we do our best to prevent breaking changes until major version releases of the provider, it is generally recommended to [pin the provider version in your configuration](https://www.terraform.io/docs/configuration/providers.html#provider-versions).

Due to the constant release pace of AWS and the relatively infrequent major version releases of the provider, there can be cases where a minor version update may contain unexpected changes depending on your configuration or environment. These may include items such as a resource requiring additional IAM permissions to support newer functionality. We typically base these decisions on a pragmatic compromise between introducing a relatively minor one-time inconvenience for a subset of the community versus better overall user experience for the entire community.

### Once a major release is published, will new features and fixes be backported to previous versions?

Generally new features and fixes will only be added to the most recent major version. Due to the high touch nature of provider development and the extensive regression testing required to ensure stability, maintaining multiple versions of the provider is not sustainable at this time. An exception to this could be a discovered security vulnerability for which backporting may be the most reasonable course of action. These would be reviewed on a case by case basis.

### AWS just announced a new region, when will I see it in the provider.

Normally pretty quickly. We usually see the region appear within the `aws-go-sdk` within a couple days of the announcement. Depending on when it lands, we can often get it out within the current or following weekly release. Comparatively, adding support for a new  region in the S3 backend can take a little longer, as it is shipped as part of Terraform Core and not via the AWS Provider.

Please note that this new region requires a manual process to enable in your account. Once enabled in the console, it takes a few minutes for everything to work properly.

If the region is not enabled properly, or the enablement process is still in progress, you may receive errors like these:

```
$ terraform apply

Error: error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid.
    status code: 403, request id: 142f947b-b2c3-11e9-9959-c11ab17bcc63

  on main.tf line 1, in provider "aws":
   1: provider "aws" {
```

To use this new region before support has been added to the Terraform AWS Provider, you can disable the provider's automatic region validation via:

```hcl
provider "aws" {
  # ... potentially other configuration ...

  region                 = "af-south-1"
  skip_region_validation = true
}

```

### How can I help?

Great question, if you have contributed before check out issues with the `help-wanted` label. These are normally enhancement issues that will have a great impact, but the maintainers are unable to develop them in the near future. If you are just getting started, take a look at issues with the `good-first-issue` label. Items with these labels will always be given priority for response.

Check out the [Contributing Guide](./README.md) for additional information.

### How can I become a maintainer?

This is an area under active research. Stay tuned!
