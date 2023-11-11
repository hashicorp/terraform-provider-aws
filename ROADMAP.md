# Roadmap:  May 2023 - July 2023

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap.

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning January to April 2023 the AWS Provider added support for the following (among many others):

- AWS VPC Lattice
- AWS Quicksight
- AWS Directory Service “Trust”
- AWS Observability Access Manager

From May - July 2023, we will be prioritizing the following areas of work:

## New Services  

### Amazon OpenSearch Serverless

Issue: [#28313](https://github.com/hashicorp/terraform-provider-aws/issues/28313)

[Amazon OpenSearch Serverless](https://aws.amazon.com/opensearch-service/features/serverless/) makes it easy for customers to run large-scale search and analytics workloads without managing clusters. It automatically provisions and scales the underlying resources to deliver fast data ingestion and query responses for even the most demanding and unpredictable workloads, eliminating the need to configure and optimize clusters.

Support for Amazon OpenSearch Serverless may include:

New Resource(s):

- `aws_opensearchserverless_collection`
- `aws_opensearchserverless_access_policy`
- `aws_opensearchserverless_security_config`
- `aws_opensearchserverless_security_policy`
- `aws_opensearchserverless_vpc_endpoint`

### AWS Clean Rooms

Issue: [#30024](https://github.com/hashicorp/terraform-provider-aws/issues/30024)

[AWS Clean Rooms](https://aws.amazon.com/clean-rooms/) helps companies and their partners more easily and securely analyze and collaborate on their collective datasets–without sharing or copying one another's underlying data. With AWS Clean Rooms, customers can create a secure data clean room in minutes, and collaborate with any other company on the AWS Cloud to generate unique insights about advertising campaigns, investment decisions, and research and development.

Support for AWS Clean Rooms may include:

New Resource(s):

- `aws_cleanrooms_collaboration`
- `aws_cleanrooms_configured_table`
- `aws_cleanrooms_configured_table_analysis_rule`
- `aws_cleanrooms_configured_table_association`
- `aws_cleanrooms_membership`

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [Lack of support for sso-session in .aws/config](https://github.com/hashicorp/terraform-provider-aws/issues/28263)
- [Cognito User Pool: cannot modify or remove schema items](https://github.com/hashicorp/terraform-provider-aws/issues/21654)
- [aws_wafv2_web_acl - Error: Provider produced inconsistent final plan](https://github.com/hashicorp/terraform-provider-aws/issues/23992)
- [aws_lb_target_group_attachment: target_id should be a list](https://github.com/hashicorp/terraform-provider-aws/issues/9901)
- [Extend Secrets Manager Rotation Configuration](https://github.com/hashicorp/terraform-provider-aws/issues/22969)

## Major Release v5

The release of version 5.0 of the Terraform AWS provider will bring highly anticipated updates to default tags, and make changes and deprecations.

### Default Tags

Default tags in the Terraform AWS provider allow practitioners to define common metadata tags at the provider level. These tags are then applied to all supported resources in the Terraform configuration. Previously, assumptions and restrictions were made to allow this feature to function across as many resources as possible. However, it could be difficult to retrofit existing code, causing frustrating manual intervention.
Thanks to new features available in the [Terraform plugin SDK](https://developer.hashicorp.com/terraform/plugin/sdkv2) and the [Terraform plugin framework](https://developer.hashicorp.com/terraform/plugin/framework), we have removed several limitations which made default tags difficult to integrate with existing resources and modules.

The updates in version 5.0 solve for:

- Inconsistent final plans that cause failures when tags are computed.
- Identical tags in both default tags and resource tags.
- Perpetual diffs within tag configurations.

### Remove EC2 Classic Functionality

In 2021 AWS [announced](https://aws.amazon.com/blogs/aws/ec2-classic-is-retiring-heres-how-to-prepare/) the retirement of EC2 Classic Networking functionality. This was scheduled to occur on August 15th, 2022. Support for the functionality was extended until late September when any AWS customers who had qualified for extension finished their migration. At that time those features were marked as deprecated and it is now time to remove them as the functionality is no longer available through AWS. While this is a standard deprecation, this is a major feature removal.

### Updating RDS Identifiers In–Place

Allow DB names to be updated in place. This is now supported by AWS, so we should allow its use. Practitioners will now be able to change names without a recreation. Details for this issue can be tracked in issue [#507](https://github.com/hashicorp/terraform-provider-aws/issues/507).

### Remove Default Value from Engine Parameters

Removes a default value that does not have a parallel with AWS and causes unexpected behavior for end users. Practitioners will now have to specify a value. Details for this issue can be tracked in issue [#27960](https://github.com/hashicorp/terraform-provider-aws/issues/27960).

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
