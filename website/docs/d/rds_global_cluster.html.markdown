---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_global_cluster"
description: |-
  Provides an RDS global cluster data source.
---

# Data Source: aws_rds_global_cluster

Provides information about an RDS global cluster.

## Example Usage

```terraform
data "aws_region" "alternate" {
	provider = "awsalternate"
}
	
data "aws_region" "current" {}

resource "aws_rds_global_cluster" "test" {
	global_cluster_identifier = "%[1]s"
	engine                    = "aurora-postgresql"
	engine_version            = "12.4"
	database_name             = "example_db"
}

resource "aws_rds_cluster" "primary" {
	engine                    = aws_rds_global_cluster.test.engine
	engine_version            = aws_rds_global_cluster.test.engine_version
	cluster_identifier        = "test-primary-cluster"
	master_username           = "username"
	master_password           = "somepass123"
	database_name             = "example_db"
	global_cluster_identifier = aws_rds_global_cluster.test.id
	db_subnet_group_name      = "test"
	skip_final_snapshot       = true
	
	depends_on = [
		aws_rds_global_cluster.test
	]
}

resource "aws_rds_cluster_instance" "primary" {
	identifier           = "test-primary-cluster-instance"
	cluster_identifier   = aws_rds_cluster.primary.id
	instance_class       = "db.r4.large"
	db_subnet_group_name = "test"
	engine               = aws_rds_global_cluster.test.engine
	engine_version       = aws_rds_global_cluster.test.engine_version
	
	depends_on = [
		aws_rds_cluster.primary
	]
}

resource "aws_rds_cluster" "secondary" {
	provider                  = "awsalternate"
	engine                    = aws_rds_global_cluster.test.engine
	engine_version            = aws_rds_global_cluster.test.engine_version
	cluster_identifier        = "test-secondary-cluster"
	global_cluster_identifier = aws_rds_global_cluster.test.id
	replication_source_identifier = aws_rds_cluster.primary.arn
	source_region 			  = data.aws_region.alternate.name
	db_subnet_group_name      = "test"
	skip_final_snapshot       = true

	depends_on = [
		aws_rds_cluster_instance.primary
	]
}

resource "aws_rds_cluster_instance" "secondary" {
	provider             = "awsalternate"
	identifier           = "test-secondary-cluster-instance"
	cluster_identifier   = aws_rds_cluster.secondary.id
	instance_class       = "db.r4.large"
	db_subnet_group_name = "test"
	engine               = aws_rds_global_cluster.test.engine
	engine_version       = aws_rds_global_cluster.test.engine_version
  
	depends_on = [
		aws_rds_cluster.secondary
	]
}

data "aws_rds_global_cluster" "test" {
	global_cluster_identifier = aws_rds_global_cluster.test.global_cluster_identifier
}
```

## Argument Reference

The following arguments are supported:

* `global_cluster_identifier` - (Required) The global cluster identifier of the RDS global cluster.

## Attributes Reference

See the [RDS Global Cluster Resource](/docs/providers/aws/r/rds_global_cluster.html) for details on the
returned attributes - they are identical.
