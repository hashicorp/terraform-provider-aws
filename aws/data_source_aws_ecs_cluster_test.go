package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEcsDataSource_ecsCluster(t *testing.T) {
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsClusterDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(dataSourceName, "pending_tasks_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "running_tasks_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "registered_container_instances_count", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
					resource.TestCheckResourceAttrPair(resourceName, "default_capacity_provider_strategy", dataSourceName, "default_capacity_provider_strategy"),
					resource.TestCheckResourceAttrPair(resourceName, "capacity_providers", dataSourceName, "capacity_providers"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "cluster_name"),
				),
			},
		},
	})
}

func TestAccAWSEcsDataSource_ecsClusterContainerInsights(t *testing.T) {
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsClusterDataSourceConfigContainerInsights,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(dataSourceName, "pending_tasks_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "running_tasks_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "registered_container_instances_count", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "setting", resourceName, "setting"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "cluster_name"),
				),
			},
		},
	})
}

var testAccCheckAwsEcsClusterDataSourceConfig = fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q
}

data "aws_ecs_cluster" "test" {
  cluster_name = "${aws_ecs_cluster.test.name}"
}
`, acctest.RandomWithPrefix("tf-acc"))

var testAccCheckAwsEcsClusterDataSourceConfigContainerInsights = fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q
  setting {
    name = "containerInsights"
    value = "enabled"
  }
}

data "aws_ecs_cluster" "test" {
  cluster_name = "${aws_ecs_cluster.test.name}"
}
`, acctest.RandomWithPrefix("tf-acc-insight"))
