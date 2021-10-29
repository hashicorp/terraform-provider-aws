---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_cluster_role_association"
description: |-
  Manages a RDS DB Cluster association with an IAM Role.
---

# Resource: aws_rds_cluster_role_association

Manages a RDS DB Cluster association with an IAM Role. Example use cases:

* [Creating an IAM Role to Allow Amazon Aurora to Access AWS Services](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Integrating.Authorizing.IAM.CreateRole.html)
* [Importing Amazon S3 Data into an RDS PostgreSQL DB Cluster](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PostgreSQL.S3Import.html)

## Example Usage

```terraform
resource "aws_rds_cluster_role_association" "example" {
  db_cluster_identifier = aws_rds_cluster.example.id
  feature_name          = "S3_INTEGRATION"
  role_arn              = aws_iam_role.example.id
}
```

## Argument Reference

The following arguments are supported:

* `db_cluster_identifier` - (Required) DB Cluster Identifier to associate with the IAM Role.
* `feature_name` - (Required) Name of the feature for association. This can be found in the AWS documentation relevant to the integration or a full list is available in the `SupportedFeatureNames` list returned by [AWS CLI rds describe-db-engine-versions](https://docs.aws.amazon.com/cli/latest/reference/rds/describe-db-engine-versions.html).
* `role_arn` - (Required) Amazon Resource Name (ARN) of the IAM Role to associate with the DB Cluster.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - DB Cluster Identifier and IAM Role ARN separated by a comma (`,`)

## Import

`aws_rds_cluster_role_association` can be imported using the DB Cluster Identifier and IAM Role ARN separated by a comma (`,`), e.g.,

```
$ terraform import aws_rds_cluster_role_association.example my-db-cluster,arn:aws:iam::123456789012:role/my-role
```
