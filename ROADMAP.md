# Roadmap:  August 2022 - October 2022

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](docs/contributing/core-services.md), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning May to July 2022, 991 Pull Requests were opened in the provider and 865 were closed/merged, adding support for the following (among many others):

- Kendra
- Transcribe
- Comprehend
- Cost Explorer
- CloudWAN

From August ‘22 - October ‘22, we will be prioritizing the following areas of work:

## New Services  

### AWS Elemental MediaLive

Issue: [#4936](https://github.com/hashicorp/terraform-provider-aws/issues/4936)

_[AWS Elemental MediaLive](https://aws.amazon.com/medialive/) is a broadcast-grade live video processing service. It lets you create high-quality video streams for delivery to broadcast televisions and internet-connected multiscreen devices, like connected TVs, tablets, smart phones, and set-top boxes._

Support for AWS Elemental MediaLive may include:

New Resource(s):

- `aws_medialive_channel`
- `aws_medialive_input`
- `aws_medialive_input_security_group`
- `aws_medialive_multiplex`
- `aws_medialive_multiplex_program`

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [Lake Formation Tag-Based Access Control](https://github.com/hashicorp/terraform-provider-aws/issues/19640)

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

### Expanding Framework based Resources

[#25606](https://github.com/hashicorp/terraform-provider-aws/pull/25606) and [#25715](https://github.com/hashicorp/terraform-provider-aws/pull/25715) added the ability for provider contributors/maintainers to implement resources and data sources based on the next generation of the provider SDK, the [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework). Over the next quarter we will increase our footprint of resources based on the framework, either by adding them as new resources or migrating existing ones. We will also ensure that resources based on the framework have parity with with provider level features, such as default tags.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
