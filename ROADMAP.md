# Roadmap:  February 2023 - April 2023

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning August to October 2022, 808 Pull Requests were opened in the provider and 783 were closed/merged, adding support for the following (among many others):

- AWS Audit Manager
- Lambda SnapStart
- RDS: Blue/Green Deployments

From February - April 2023, we will be prioritizing the following areas of work:

## New Services  

### AWS Quicksight

Issue: [#10990]([https://github.com/hashicorp/terraform-provider-aws/issues/17981](https://github.com/hashicorp/terraform-provider-aws/issues/10990))

[AWS Quicksight](https://aws.amazon.com/quicksight/) has a serverless architecture that automatically scales to hundreds of thousands of users without the need to set up, configure, or manage your own servers. It also ensures that your users don’t have to deal with slow dashboards during peak hours, when multiple business intelligence (BI) users are accessing the same dashboards or datasets. And with pay-per-session pricing, you pay only when your users access the dashboards or reports, which makes it cost effective for deployments with many users. QuickSight is also built with robust security, governance, and global collaboration features for your enterprise workloads.

Support for AWS Quicksight may include:

New Resource(s):

- `aws_quicksight_iam_policy_assignment`
- `aws_quicksight_data_set`
- `aws_quicksight_ingestion`
- `aws_quicksight_template`
- `aws_quicksight_dashboard`
- `aws_quicksight_template_alias`

### AWS Recycle Bin

Issue: [#23160](https://github.com/hashicorp/terraform-provider-aws/issues/23160)

[AWS Recycle Bin](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/recycle-bin.html) is a data recovery feature that enables you to restore accidentally deleted Amazon EBS snapshots and EBS-backed AMIs. When using Recycle Bin, if your resources are deleted, they are retained in the Recycle Bin for a time period that you specify before being permanently deleted.

Support for AWS Recycle Bin may include:

New Resource(s):

- `aws_recycle_bin_rule`

### AWS Directory Service “Trust”

Issue: [#11901](https://github.com/hashicorp/terraform-provider-aws/issues/11901)

Easily integrate AWS Managed Microsoft AD with your existing AD by using AD trust relationships. Using trusts enables you to use your existing Active Directory to control which AD users can access your AWS resources.

Support for AWS Director Service "Trust" may include:

New Resource(s):

- `aws_directory_service_directory_trust`

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [Resource Identifiers and Tags for VPC Security Group Rules](https://github.com/hashicorp/terraform-provider-aws/issues/20104)
- [Better Lambda error](https://github.com/hashicorp/terraform-provider-aws/issues/13709)
- [AssumeRoleTokenProviderNotSetError when using assume_role with mfa enabled](https://github.com/hashicorp/terraform-provider-aws/issues/10491)
- [Proposal: Add support Object-level logging in the existing trail for resource 'aws_s3_bucket'](https://github.com/hashicorp/terraform-provider-aws/issues/9459)
- [Proposal: Add support Object-level logging in the existing trail for resource 'aws_s3_bucket'](https://github.com/hashicorp/terraform-provider-aws/issues/9459)
- [Add support for elasticsearch outbound connection and relevant accepter](https://github.com/hashicorp/terraform-provider-aws/pull/22988)
- [Add support for Route 53 IP Based Routing Policy](https://github.com/hashicorp/terraform-provider-aws/issues/25321)
- [Add ability to query ECR repository for most recently pushed image](https://github.com/hashicorp/terraform-provider-aws/issues/12798)

### Default Tags

[#17829](https://github.com/hashicorp/terraform-provider-aws/issues/17829) added the `default_tags` block to allow practitioners to tags at the provider level. This allows configured resources capable of assigning tags to have them inherit those as well as be able to specify them at the resource level. This has proven extremely popular with the community, however it comes with a number of significant caveats ([#18311](https://github.com/hashicorp/terraform-provider-aws/issues/18311), [#19583](https://github.com/hashicorp/terraform-provider-aws/issues/19583), [#19204](https://github.com/hashicorp/terraform-provider-aws/issues/19204)) for use which have resulted from limitations in the provider SDK we use. New functionality in the [terraform-plugin-sdk](https://github.com/hashicorp/terraform-plugin-sdk) and [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework) should allow us to temper these caveats. This quarter we plan to begin the development of this feature, based on the research completed last quarter by the engineering team.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
