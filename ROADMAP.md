# Roadmap:  August 2021 - October 2021

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top 10 Community Issues, [Core Services](docs/contributing/core-services.md), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur .

In the period spanning May to July 2021 539 Pull Requests were opened in the provider and 449 were merged, adding support for:

- Amazon Timestream
- AWS AppConfig
- AWS Amplify
- AWS Service Catalog
- AWS ElasticSearch Native SAML for Kibana
- Amazon Macie 2
- Delegated Administrators for Organisations
- Predictive Autoscaling
- Amazon EKS OIDC
- AWS Transfer Family support for Amazon Elastic File System
- Amazon Kinesis Data Streams for Amazon DynamoDB

Among many other enhancements, bug fixes and resolutions to technical debt items.

From August-October ‘21, we will be prioritizing the following areas of work:

## Provider Version v4.0.0

Issue: [#20433](https://github.com/hashicorp/terraform-provider-aws/issues/20433)

The next major release of the provider will include the adoption of the AWS Go SDK v2.0 as well as a refactor of one of our oldest and most used resources: S3.

There will also be the usual deprecations and sometimes breaking changes to existing resources which are necessary to maintain consistency of behavior across resources. Our goal is to focus on standardization to reduce technical debt and lay a strong foundation for future enhancement initiatives within the provider.

For details of the changes in full please refer to #20433. We would love to hear your feedback.

## New Services

### Amazon Quicksight
Issue: [#10990](https://github.com/hashicorp/terraform-provider-aws/issues/10990)

_Amazon QuickSight is a scalable, serverless, embeddable, machine learning-powered business intelligence (BI) service built for the cloud. QuickSight lets you easily create and publish interactive BI dashboards that include Machine Learning-powered insights. QuickSight dashboards can be accessed from any device, and seamlessly embedded into your applications, portals, and websites._

Support for Amazon Quicksight will include:

New Resource(s):
- aws_quicksight_data_source
- aws_quicksight_group_membership
- aws_quicksight_iam_policy_assignment
- aws_quicksight_data_set
- aws_quicksight_ingestion
- aws_quicksight_template
- aws_quicksight_dashboard
- aws_quicksight_template_alias


### Amazon AppStream
Issue: [#6058](https://github.com/hashicorp/terraform-provider-aws/issues/6508)

_Amazon AppStream 2.0 is a fully managed non-persistent desktop and application virtualization service that allows your users to securely access the data, applications, and resources they need, anywhere, anytime, from any supported device. With AppStream 2.0, you can easily scale your applications and desktops to any number of users across the globe without acquiring, provisioning, and operating hardware or infrastructure. AppStream 2.0 is built on AWS, so you benefit from a data center and network architecture designed for the most security-sensitive organizations. Each end user has a fluid and responsive experience because your applications run on virtual machines optimized for specific use cases and each streaming sessions automatically adjusts to network conditions._

Support for Amazon AppStream will include:

New Resource(s):
- aws_appstream_stack
- aws_appstream_fleet
- aws_appstream_imagebuilder

### Amazon Connect 
Issue: [#16392](https://github.com/hashicorp/terraform-provider-aws/issues/16392)

_Amazon Connect is an easy to use omnichannel cloud contact center that helps you provide superior customer service at a lower cost. Designed from the ground up to be omnichannel, Amazon Connect provides a seamless experience across voice and chat for your customers and agents. This includes one set of tools for skills-based routing, task management, powerful real-time and historical analytics, and intuitive management tools – all with pay-as-you-go pricing, which means Amazon Connect simplifies contact center operations, improves agent efficiency, and lowers costs. You can set up a contact center in minutes that can scale to support millions of customers from the office or as a virtual contact center._

Support for Amazon Connect will include:

New Resource(s):
- aws_connect_instance
- aws_connect_contact_flow
- aws_connect_bot_association
- aws_connect_lex_bot_association
- aws_connect_lambda_function_association

New Data Source(s): 
- aws_connect_instance
- aws_connect_contact_flow
- aws_connect_bot_association
- aws_connect_lex_bot_association
- aws_connect_lambda_function_association

## Enhancements to Existing Services
- [Support for KMS Multi-Region Keys](https://github.com/hashicorp/terraform-provider-aws/issues/19896)
- [S3 Replication Time Control](https://github.com/hashicorp/terraform-provider-aws/issues/10974)
- [New Data Source: aws_iam_roles](https://github.com/hashicorp/terraform-provider-aws/issues/14470)

## Project Restructure: Service Level Packages

The scale of the provider (now 1000 resources/datasources) has led to its existing package structure being difficult to work with and maintain. This quarter we are going to perform a large refactor of the codebase, to align on a single go package per AWS service. More details can be found in the encompassing issue [#20431](https://github.com/hashicorp/terraform-provider-aws/issues/20431)

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

### Scaffolding for new resources, datasources and associated tests. 

Adding resources, datasources and test files to the provider is a repetitive task which should be automated to ensure consistency and speed up contributor and maintainer workflow. A simple cli tool should be able to generate these files in place, and ensure that any code reference additions required (ie adding to `provider.go`) are performed as part of the process.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
