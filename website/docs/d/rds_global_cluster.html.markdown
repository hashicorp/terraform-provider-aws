---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_global_cluster"
description: |-
  Terraform data source for managing an AWS RDS (Relational Database) Global Cluster.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->

# Data Source: aws_rds_global_cluster

Terraform data source for managing an AWS RDS (Relational Database) Global Cluster.

## Example Usage

### Basic Usage

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

The following arguments are required:

* `global_cluster_identifier` - (Required) The global cluster identifier of the RDS global cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - RDS Global Cluster Amazon Resource Name (ARN)
* `global_cluster_identifier` - Global cluster identifier.
* `database_name` - Name of the automatically created database on cluster creation.
* `deletion_protection` -  If the Global Cluster should have deletion protection enabled. The database can't be deleted when this value is set to `true`.
* `engine` - Name of the database engine.
* `engine_version` -   Version of the database engine for this Global Cluster.
* `storage_encrypted` - Whether the DB cluster is encrypted.
* `global_cluster_members` -  Set of objects containing Global Cluster members.
  * `db_cluster_arn` - Amazon Resource Name (ARN) of member DB Cluster
  * `is_writer` - Whether the member is the primary DB Cluster
* `global_cluster_resource_id` - AWS Region-unique, immutable identifier for the global database cluster. 


		
					resource.TestCheckResourceAttrPair(dataSourceName, "deletion_protection", resourceName, "deletion_protection"),


					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_members", resourceName, "global_cluster_members"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_resource_id", resourceName, "global_cluster_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),