# Roadmap:  February 2022 - April 2022

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](../contributing/core-services.md), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning November 2021 to Janury 2022, 668 Pull Requests were opened in the provider and 730 were closed/merged, adding support for:

- Managing Amazon CloudSearch Domains
- ECS Task Sets
- S3 Intelligent Tiering Archive Configuration
- IoT Thing Group
- Lambda Access Points for S3
- ECR Public Repositories

From February ‘22 - April ‘22, we will be prioritizing the following areas of work:

## New Services  

### AWS AppFlow

Issue: [#16253](https://github.com/hashicorp/terraform-provider-aws/issues/16253)

_[Amazon AppFlow](https://aws.amazon.com/appflow/) is a fully managed integration service that enables you to securely transfer data between Software-as-a-Service (SaaS) applications_

Support for Amazon AppFlow will include:

New Resource(s):

- `aws_appflow_flow`
- `aws_appflow_connector_profile`


### Amazon Global Networks

Issue: [#11132](https://github.com/hashicorp/terraform-provider-aws/issues/11132)

_A [Global Network](https://docs.aws.amazon.com/vpc/latest/tgwnm/global-networks.html) is a container for your network objects. After you create it, you can register your transit gateways and define your on-premises networks in the global network._

Support for Global Networks will include:

New Resource(s):

- `aws_networkmanager_global_network`
- `aws_networkmanager_site`
- `aws_networkmanager_link`
- `aws_networkmanager_device`
- `aws_networkmanager_transit_gateway_registration`

### Amazon OpenSearch Service

Issue: [#20853](https://github.com/hashicorp/terraform-provider-aws/issues/20853)

_[Amazon OpenSearch Service](https://aws.amazon.com/opensearch-service/) is a distributed, open-source search and analytics suite used for a broad set of use cases like real-time application monitoring, log analytics, and website search_

Affected Resource(s):

- `aws_elasticsearch_domain`

### Amazon Managed Grafana

Issue: [#16789](https://github.com/hashicorp/terraform-provider-aws/issues/16789)

_[Amazon Managed Grafana](https://aws.amazon.com/grafana) is a fully managed service for open source Grafana that enables you to query, visualize, alert on and understand your metrics._

Support for Amazon Managed Grafana will include:

New Resource(s):

- `aws_grafana_workspace`
- `aws_grafana_license_association`

## Enhancements to Existing Services

- [changing identifier in RDS (aws_db_instance) will destroy/create the db](https://github.com/hashicorp/terraform-provider-aws/issues/507)
- [Terraform fails to destroy autoscaling group if scale in protection is enabled](https://github.com/hashicorp/terraform-provider-aws/issues/5278)
- [Implement support for time based retention policies in DLM](https://github.com/hashicorp/terraform-provider-aws/issues/11456)
- [Cannot destroy attached security groups](https://github.com/hashicorp/terraform-provider-aws/issues/13593)
- [default_tags always shows an update](https://github.com/hashicorp/terraform-provider-aws/issues/18311)
- [AWS Synthetics Canary Missing support for Environment Variables](https://github.com/hashicorp/terraform-provider-aws/issues/17948)
- [Add retry handling when a request's connection is reset by peer](https://github.com/hashicorp/terraform-provider-aws/issues/10715)
- [resource/aws_db_instance: Should support enabling cross-region automated backups](https://github.com/hashicorp/terraform-provider-aws/issues/16708)
- [Feature Request: Support Route53 Domains](https://github.com/hashicorp/terraform-provider-aws/issues/88)
- [import aws_s3_bucket does not store important attributes like acl](https://github.com/hashicorp/terraform-provider-aws/issues/6193)
- [Add user_group argument to aws_elasticache_replication_group](https://github.com/hashicorp/terraform-provider-aws/issues/20328)
- [aws_lb_listener_certificate not destroyed upon "force new resource"](https://github.com/hashicorp/terraform-provider-aws/issues/7761)
- [AWS ACM Expected certificate to be issued but was in state PENDING_VALIDATION](https://github.com/hashicorp/terraform-provider-aws/issues/9338)
- [Feature Request: Dynamic Security Group Association for VPC Endpoint Interface](https://github.com/hashicorp/terraform-provider-aws/issues/10429)
- [Inconsistency in AWS Terraform provider with aws_lambda_function](https://github.com/hashicorp/terraform-provider-aws/issues/11787)
- [aws_wafv2_web_acl: managed-rule-group-statement is missing Version option](https://github.com/hashicorp/terraform-provider-aws/issues/21546)
- [Transit Gateway multicast support](https://github.com/hashicorp/terraform-provider-aws/issues/11120)
- [S3 bucket slow to delete when destroyed during an apply](https://github.com/hashicorp/terraform-provider-aws/issues/12146)
- [aws_elasticsearch_domain cognito_options cause Cycle Error](https://github.com/hashicorp/terraform-provider-aws/issues/5557)
- [data source for aws_iam_saml_provider](https://github.com/hashicorp/terraform-provider-aws/issues/7283)
- [Amazon MSK multiple authentication modes and updates to TLS encryption settings](https://github.com/hashicorp/terraform-provider-aws/issues/20956)
- [Support cold storage option for aws_elasticsearch_domain config](https://github.com/hashicorp/terraform-provider-aws/issues/19593)
- [aws_sns_topic_subscription doesn't support unconfirmed endpoints](https://github.com/hashicorp/terraform-provider-aws/issues/7782)
- [Add Radius support for aws_directory_service_directory](https://github.com/hashicorp/terraform-provider-aws/issues/12639)
- [New resource `aws_kafkaconnect_connector`](https://github.com/hashicorp/terraform-provider-aws/issues/20942)
- [aws_rds_cluster_activity_stream](https://github.com/hashicorp/terraform-provider-aws/pull/22097)
- [Amazon EMR on Amazon EKS](https://github.com/hashicorp/terraform-provider-aws/issues/16717)
- [Add destination metrics for Replication rule](https://github.com/hashicorp/terraform-provider-aws/issues/16347)
- [add domain validation options parameter to aws_acm_certificate](https://github.com/hashicorp/terraform-provider-aws/issues/3851)
- [Support Starting AWS Database Migration Service Replication Task](https://github.com/hashicorp/terraform-provider-aws/issues/2083)
- [Feature Request: DynamoDB CloudWatch Contributor Insights](https://github.com/hashicorp/terraform-provider-aws/issues/13933)
- [Feature request: Support Api Gateway Canary release](https://github.com/hashicorp/terraform-provider-aws/issues/2727)
- [DMS task modification lifecycle](https://github.com/hashicorp/terraform-provider-aws/issues/2236)

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

### Scaffolding for new resources, datasources and associated tests

Adding resources, datasources and test files to the provider is a repetitive task which should be automated to ensure consistency and speed up contributor and maintainer workflow. A simple cli tool should be able to generate these files in place, and ensure that any code reference additions required (ie adding to `provider.go`) are performed as part of the process.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
