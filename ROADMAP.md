# Roadmap:  May 2022 - July 2022

Every few months, the team will highlight areas of focus for our work and upcoming research.

We select items for inclusion in the roadmap from the Top Community Issues, [Core Services](docs/contributing/core-services.md), and internal priorities. Where community sourced contributions exist we will work with the authors to review and merge their work. Where this does not exist or the original contributors are not available we will create the resources and implementation ourselves.

Each weekly release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. To view all the items we've prioritized for this quarter, please see the [Roadmap milestone](https://github.com/hashicorp/terraform-provider-aws/milestone/138).

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

In the period spanning February to April 2022, 912 Pull Requests were opened in the provider and 839 were closed/merged, adding support for:

- Amazon OpenSearch
- AWS Cost Categories
- AWS AppFlow
- Amazon Managed Grafana
- Amazon Global Networks
- Lambda Function URLs

From May ‘22 - July ‘22, we will be prioritizing the following areas of work:

## New Services  

### Amazon Transcribe & Transcribe Medical

Issue: [#18865](https://github.com/hashicorp/terraform-provider-aws/issues/18865)

_[Amazon Transcribe](https://aws.amazon.com/transcribe/) is an automatic speech recognition service that makes it easy to add speech to text capabilities to any application. Transcribe’s features enable you to ingest audio input, produce easy to read and review transcripts, improve accuracy with customization, and filter content to ensure customer privacy._

Support for Amazon Transcribe will include:

New Resource(s):

- `aws_transcribe_language_model`
- `aws_transcribe_medical_vocabulary`
- `aws_transcribe_vocabulary`
- `aws_transcribe_vocabulary_filter`


### Amazon Comprehend & Comprehend Medical

Issue: [#18864](https://github.com/hashicorp/terraform-provider-aws/issues/18864)

_[Amazon Comprehend](https://aws.amazon.com/comprehend/) is a natural-language processing (NLP) service that uses machine learning to uncover valuable insights and connections in text._

Support for Amazon Comprehend will include:

New Resource(s):

- `aws_comprehend_endpoint`
- `aws_comprehend_entity_recognizer`
- `aws_comprehend_document_classifier`

### Amazon Textract

Issue: [#24478](https://github.com/hashicorp/terraform-provider-aws/issues/24478)

_[Amazon Textract](https://aws.amazon.com/textract/) is a machine learning (ML) service that automatically extracts text, handwriting, and data from scanned documents._

New Resource(s):

- TBD

### Amazon Kendra

Issue: [#13367](https://github.com/hashicorp/terraform-provider-aws/issues/13367)

_[Amazon Kendra](https://aws.amazon.com/kendra/) is an intelligent search service powered by machine learning (ML)._

Support for Amazon Kendra will include:

New Resource(s):

- `aws_kendra_index`
- `aws_kendra_query_suggestion_block_list`
- `aws_kendra_thesaurus`

## Enhancements to Existing Services

- [Lake Formation Tag-Based Access Control](https://github.com/hashicorp/terraform-provider-aws/issues/19640)
- [Amazon Managed Apache Cassandra Service / Keyspaces](https://github.com/hashicorp/terraform-provider-aws/issues/11221)
- [Assignment multiple users or groups via aws_ssoadmin_account_assignment](https://github.com/hashicorp/terraform-provider-aws/issues/18739)
- [Add support for regex_match_statement to AWS WAF v2 ACL rules](https://github.com/hashicorp/terraform-provider-aws/pull/22452)
- [Introduce custom timeout when waiting for aws_ecs_service to reach a steady state](https://github.com/hashicorp/terraform-provider-aws/pull/18868)
- [r/aws_wafv2_web_acl: add support for captcha in rule actions #21766](https://github.com/hashicorp/terraform-provider-aws/pull/21766)
- [Terraform seems to ignore "skip_final_snapshot" for rds cluster](https://github.com/hashicorp/terraform-provider-aws/issues/2588)
- [Provider produced inconsistent final plan / an invalid new value for .tags_all](https://github.com/hashicorp/terraform-provider-aws/issues/19583)
- [Support for Reserved Instances](https://github.com/hashicorp/terraform-provider-aws/issues/8521)
- [Cost Explorer](https://github.com/hashicorp/terraform-provider-aws/issues/16137)
- [Directory Service Shared Directory Support](https://github.com/hashicorp/terraform-provider-aws/issues/6004)
- [New Feature: Launch AWS Marketplace products in linked AWS accounts](https://github.com/hashicorp/terraform-provider-aws/issues/17146)
- [Don't mark non SecureString SSM parameters as sensitive](https://github.com/hashicorp/terraform-provider-aws/issues/9090)
- [FIS Experiment Template](https://github.com/hashicorp/terraform-provider-aws/issues/18125)
- [AWS Inspector Enable Service Feature](https://github.com/hashicorp/terraform-provider-aws/issues/22330)
- [Fix aws_iam_policy_document order](https://github.com/hashicorp/terraform-provider-aws/pull/23060)
- [Need SSM update-service-setting equivalent](https://github.com/hashicorp/terraform-provider-aws/pull/13018)
- [Add data sources for Managed Rules for WAF and WAF Regional](https://github.com/hashicorp/terraform-provider-aws/pull/10563)
- [Add aws_acmpca_permission resource](https://github.com/hashicorp/terraform-provider-aws/pull/12485)
- [Terraform AWS Provider does not support CopyDBSnapshot](https://github.com/hashicorp/terraform-provider-aws/issues/9885)
- [Access Analyzer archive_rule](https://github.com/hashicorp/terraform-provider-aws/issues/11102)
- [Add new resource_aws_lightsail_container_service](https://github.com/hashicorp/terraform-provider-aws/pull/20625)
- [Add resource aws_db_snapshot_copy](https://github.com/hashicorp/terraform-provider-aws/pull/9886)
- [Dualstack prefix forcibly removed from ALIAS records](https://github.com/hashicorp/terraform-provider-aws/issues/6480)
- [Add LoadBasedAutoscaling to OpsWorks Layer](https://github.com/hashicorp/terraform-provider-aws/pull/10962)
- [Allow setting custom domain for click/open tracking as part of SES Event destination resource](https://github.com/hashicorp/terraform-provider-aws/issues/6339)
- [Aws_inspector_assessment_template ability to send findings to SNS topic](https://github.com/hashicorp/terraform-provider-aws/issues/843)
- [data_source/aws_ecr_lifecycle_policy_document: adding new data source for ECR](https://github.com/hashicorp/terraform-provider-aws/pull/6133)

## Research Topics

Research topics include features, architectural changes, and ideas that we are pursuing in the longer term that may significantly impact the core user experience of the AWS provider. Research topics are discovery only and are not guaranteed to be included in a future release.

### Mux with Plugin Framework 

Pull Request: [#23948](https://github.com/hashicorp/terraform-provider-aws/pull/23948)

Prototyping the ability to [combine](https://www.terraform.io/plugin/mux) resources and data sources implemented in either Terraform Plugin SDK v2 or Terraform Plugin Framework using `terraform-plugin-mux`.

### AWS Cloud Control Community Documentation 

Issue: [#469](https://github.com/hashicorp/terraform-provider-awscc/issues/469)

Currently the `awscc` provider documentation that lives on the registry is generated from the CloudFormation Cloud Control API schema. This means that we are limited to attribute level and resource level descriptions that are quite terse. The `aws` Provider has rich, contributor drafted documentation which includes examples, notes, and guides that make for a much more positive user experience.

While resource behavior in `awscc` should remain wholly generated, we would like to enable contributors to append information to the generated documentation in order to foster an improved experience more inline with what Terraform practitioners are used to.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.
