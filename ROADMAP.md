# Roadmap:  November 2022 - January 2022

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning August to October 2022, 808 Pull Requests were opened in the provider and 783 were closed/merged, adding support for the following (among many others):

- Elemental MediaLive
- RDS Reserved Instances
- SESV2
- BYOIP Support for Global Accelerator
- RDS Custom Managed Databases

From November ‘22 - January ‘23, we will be prioritizing the following areas of work:

## New Services  

### AWS Audit Manager

Issue: [#4936](https://github.com/hashicorp/terraform-provider-aws/issues/17981)

_[AWS Audit Manager](https://aws.amazon.com/audit-manager/) helps you continuously audit your AWS usage to simplify how you assess risk and compliance with regulations and industry standards. Audit Manager automates evidence collection to reduce the “all hands on deck” manual effort that often happens for audits and enable you to scale your audit capability in the cloud as your business grows. With Audit Manager, it is easy to assess if your policies, procedures, and activities – also known as controls – are operating effectively. When it is time for an audit, AWS Audit Manager helps you manage stakeholder reviews of your controls and enables you to build audit-ready reports with much less manual effort._

Support for AWS Audit Manager may include:

New Resource(s):

- `aws_auditmanager_assessment`
- `aws_auditmanager_assessment_framework`
- `aws_auditmanager_assessment_report`
- `aws_auditmanager_control`

## Re:Invent

This quarter includes Re:Invent 2022 and as such, we will be looking out for key launch events and aligning provider support to those most desired by the community.

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [SSO: Support for permission boundary policies](https://github.com/hashicorp/terraform-provider-aws/issues/25893)
- [Stop Instances](https://github.com/hashicorp/terraform-provider-aws/issues/22)
- [Proposal: Add support Object-level logging in the existing trail for resource 'aws_s3_bucket'](https://github.com/hashicorp/terraform-provider-aws/issues/9459)
- [Add support for auto-adjusting budgets](https://github.com/hashicorp/terraform-provider-aws/issues/23268)
- [Support managed rule group configs in aws_wafv2_web_acl for the new managed rule AWSManagedRulesATPRuleSet](https://github.com/hashicorp/terraform-provider-aws/issues/23290)
- [Add Support for WAFv2 Managed Rule Group Configuration](https://github.com/hashicorp/terraform-provider-aws/issues/23287)
- [Add data source aws_organizations_accounts](https://github.com/hashicorp/terraform-provider-aws/pull/18589)
- [Add support for elasticsearch outbound connection and relevant accepter](https://github.com/hashicorp/terraform-provider-aws/pull/22988)
- [provider: Add scrubbing for sensitive data in logs](https://github.com/hashicorp/terraform-provider-aws/issues/26029)
- [Add support for Route 53 IP Based Routing Policy](https://github.com/hashicorp/terraform-provider-aws/issues/25321)
- [Add ability to query ECR repository for most recently pushed image](https://github.com/hashicorp/terraform-provider-aws/issues/12798)
- [rds: export db snapshot data to S3](https://github.com/hashicorp/terraform-provider-aws/issues/16181)

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

### Default Tags

[#17829](https://github.com/hashicorp/terraform-provider-aws/issues/17829) added the `default_tags` block to allow practitioners to tags at the provider level. This allows configured resources capable of assigning tags to have them inherit those as well as be able to specify them at the resource level. This has proven extremely popular with the community, however it comes with a number of significant caveats ([#18311](https://github.com/hashicorp/terraform-provider-aws/issues/18311), [#19583](https://github.com/hashicorp/terraform-provider-aws/issues/19583), [#19204](https://github.com/hashicorp/terraform-provider-aws/issues/19204)) for use which have resulted from limitations in the provider SDK we use. New functionality in the [terraform-plugin-sdk](https://github.com/hashicorp/terraform-plugin-sdk) and [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework) should allow us to temper these caveats. This quarter we plan to research and propose changes to allow `default_tags` to be more widely usable.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
