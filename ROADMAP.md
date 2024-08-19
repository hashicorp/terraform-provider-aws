# Roadmap:  Feb 2024 - Apr 2024

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](https://hashicorp.github.io/terraform-provider-aws/core-services/), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap.

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning Nov to Jan 2024 the AWS Provider added support for the following (among many others):

- Amazon S3 Express
- Amazon S3 Access Controls
- Amazon DocDB Elastic Cluster
- Amazon EBS Fast Snapshot Restore
- Amazon Bedrock

From Feb - April 2024, we will be prioritizing the following areas of work:

## New Services

### AWS Resource Explorer Search

Issue: [#36033](https://github.com/hashicorp/terraform-provider-aws/issues/36033)

[Resource Explorer](https://aws.amazon.com/resourceexplorer/) Use AWS Resource Explorer to more easily search for and discover your resources across AWS Regions and accounts, such as Amazon Elastic Compute Cloud (Amazon EC2) instances, Amazon Kinesis streams, and Amazon DynamoDB tables..

Support for additional Resource explorer resources may include:

New Resource(s):

- `aws_resourceexplorer2_search`

### Amazon Verified Permissions

Issue: [#32158](https://github.com/hashicorp/terraform-provider-aws/issues/32158)

[Amazon Verified Permissions](https://aws.amazon.com/verified-permissions/) helps developers build more secure applications faster by externalizing authorization and centralizing policy management. They can also align application access with Zero Trust principles.

Support for Amazon Verified Permissions may include:

New Resource(s):

- `aws_verifiedpermissions_policy`
- `aws_verifiedpermissions_identity_source`

### Amazon Security Lake

Issue: [#29376](https://github.com/hashicorp/terraform-provider-aws/issues/29376)

[Amazon Security Lake](https://aws.amazon.com/security-lake/) automatically centralizes security data from AWS environments, SaaS providers, on premises, and cloud sources into a purpose-built data lake stored in your account. With Security Lake, you can get a more complete understanding of your security data across your entire organization. You can also improve the protection of your workloads, applications, and data.

Support for Amazon Security Lake may include:

New Resource(s):

- `aws_security_lake_aws_log_source`
- `aws_security_lake_custom_log_source`
- `aws_security_lake_subscriber`

### Amazon DevOps Guru

Issue: [#17919](https://github.com/hashicorp/terraform-provider-aws/issues/17919)

[Amazon DevOps Guru](https://aws.amazon.com/security-lake/) uses ML to detect abnormal operating patterns so you can identify operational issues before they impact your customers.

Support for Amazon DevOps Guru may include:

New Resource(s):

- `aws_devopsguru_notification_channel`
- `aws_devopsguru_resource_collection`

## Enhancements to Existing Services

This quarter most of our efforts will be focused on enhancements and stability improvements of our core services, rather than adding brand new services to the provider. The following list comprises the items most important to the community.

- [Add in Security Hub Automation Rules](https://github.com/hashicorp/terraform-provider-aws/issues/32210)
- [aws rds modify-certificates](https://github.com/hashicorp/terraform-provider-aws/issues/33196)
- [Add EKS cluster IAM access management API resources](https://github.com/hashicorp/terraform-provider-aws/issues/34982)
- [Support for AWS Shield Advance Subscriptions](https://github.com/hashicorp/terraform-provider-aws/issues/21430)
- [Add resources for ComputeOptimizer Recommendation Preferences](https://github.com/hashicorp/terraform-provider-aws/issues/23945)

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
