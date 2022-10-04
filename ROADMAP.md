# Roadmap:  August 2022 - October 2022

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

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

### AWS Audit Manager

Issue: [#4936](https://github.com/hashicorp/terraform-provider-aws/issues/17981)

_[AWS Audit Manager](https://aws.amazon.com/audit-manager/) helps you continuously audit your AWS usage to simplify how you assess risk and compliance with regulations and industry standards. Audit Manager automates evidence collection to reduce the “all hands on deck” manual effort that often happens for audits and enable you to scale your audit capability in the cloud as your business grows. With Audit Manager, it is easy to assess if your policies, procedures, and activities – also known as controls – are operating effectively. When it is time for an audit, AWS Audit Manager helps you manage stakeholder reviews of your controls and enables you to build audit-ready reports with much less manual effort._

Support for AWS Audit Manager may include:

New Resource(s):

- `aws_auditmanager_assessment`
- `aws_auditmanager_assessment_framework`
- `aws_auditmanager_assessment_report`
- `aws_auditmanager_control`

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [Add support for AWS MSK Serverless cluster type](https://github.com/hashicorp/terraform-provider-aws/issues/22058)
- [VPC Reachability Analyzer / EC2 Network Insights](https://github.com/hashicorp/terraform-provider-aws/issues/16715)
- [Amazon S3 Storage Lens configuration](https://github.com/hashicorp/terraform-provider-aws/issues/16310)
- [CodePipeline: ECR as Source Action](https://github.com/hashicorp/terraform-provider-aws/issues/7012)
- [Resource Identifiers and Tags for VPC Security Group Rules](https://github.com/hashicorp/terraform-provider-aws/issues/20104)
- [Add support for regex_match_statement to AWS WAF v2 ACL rules](https://github.com/hashicorp/terraform-provider-aws/issues/22452)
- [Redact Sensitive Variables from Logging](https://github.com/hashicorp/terraform-provider-aws/issues/26029)
- [Support for RDS Reserved Instances](https://github.com/hashicorp/terraform-provider-aws/issues/8521)
- [New Feature: Launch AWS Marketplace produccts (AMIs, containers) in linked AWS accounts](https://github.com/hashicorp/terraform-provider-aws/issues/17146)
- [AWS Inspector2 Enable Service Feature](https://github.com/hashicorp/terraform-provider-aws/issues/22330)
- [aws_sns_platform_application: support APNS with token-based authentication](https://github.com/hashicorp/terraform-provider-aws/issues/23147)
- [Cannot use SQS redrive_allow_policy correctly without creating a cycle](https://github.com/hashicorp/terraform-provider-aws/issues/22577)
- [aws_sns_platform_application: support APNS with token-based authentication](https://github.com/hashicorp/terraform-provider-aws/issues/23147)
- [dms-vpc-role is not configured properly when creating aws_dms_replication_instance](https://github.com/hashicorp/terraform-provider-aws/issues/11025)
- [Modify aws_db_instance and delete aws_db_parameter_group breaks](https://github.com/hashicorp/terraform-provider-aws/issues/6448)
- [Add support for setting default SSM patch baseline](https://github.com/hashicorp/terraform-provider-aws/issues/3342)
- [Add force_delete to aws_backup_vault resource](https://github.com/hashicorp/terraform-provider-aws/issues/13247)
- [Support of dedicated IP pool in AWS SES](https://github.com/hashicorp/terraform-provider-aws/issues/10703)
- [Do not try to delete lambda@edge functions with replicas](https://github.com/hashicorp/terraform-provider-aws/issues/1721)
- [Terraform seems to ignore "skip_final_snapshot" for rds cluster](https://github.com/hashicorp/terraform-provider-aws/issues/2588)
- [Cognito User Pool: cannot modify or remove schema items](https://github.com/hashicorp/terraform-provider-aws/issues/21654)
- [Support for SES domain and email identity default configuration set](https://github.com/hashicorp/terraform-provider-aws/issues/21129)


## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

### Expanding Framework based Resources

[#25606](https://github.com/hashicorp/terraform-provider-aws/pull/25606) and [#25715](https://github.com/hashicorp/terraform-provider-aws/pull/25715) added the ability for provider contributors/maintainers to implement resources and data sources based on the next generation of the provider SDK, the [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework). Over the next quarter we will increase our footprint of resources based on the framework, either by adding them as new resources or migrating existing ones. We will also ensure that resources based on the framework have parity with with provider level features, such as default tags. Beginning this migration will give us access to new features and functionality in the framework, enabling an improved experience in the framework based resources.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
