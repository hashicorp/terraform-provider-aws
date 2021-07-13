package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDataSourceAWSRDSGlobalCluster_basic(t *testing.T) {
	var providers []*schema.Provider
	globalClusterName := fmt.Sprintf("testaccawsrdsglobalcluster-basic-%s", acctest.RandString(10))
	dataSourceName := "data.aws_rds_global_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccAlternateAccountPreCheck(t)
		},
		ErrorCheck:        testAccErrorCheck(t, rds.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRdsGlobalClusterConfigBasic(globalClusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "database_name", resourceName, "database_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "deletion_protection", resourceName, "deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_identifier", resourceName, "global_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_members", resourceName, "global_cluster_members"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_resource_id", resourceName, "global_cluster_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRdsGlobalClusterConfigBasic(globalClusterName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(2),
		fmt.Sprintf(`

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
`, globalClusterName))
}
