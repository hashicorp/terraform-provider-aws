package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSMskClusterDataSource_Name(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_msk_cluster.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterDataSourceConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "bootstrap_brokers"),
					resource.TestCheckResourceAttrSet(dataSourceName, "bootstrap_brokers_tls"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", dataSourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(resourceName, "kafka_version", dataSourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(resourceName, "number_of_broker_nodes", dataSourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "zookeeper_connect_string", dataSourceName, "zookeeper_connect_string"),
				),
			},
		},
	})
}

func testAccMskClusterDataSourceConfigName(rName string) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.1.0"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = ["${aws_subnet.example_subnet_az1.id}", "${aws_subnet.example_subnet_az2.id}", "${aws_subnet.example_subnet_az3.id}"]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = ["${aws_security_group.example_sg.id}"]
  }

  tags = {
    foo = "bar"
  }
}

data "aws_msk_cluster" "test" {
  cluster_name = "${aws_msk_cluster.test.cluster_name}"
}
`, rName)
}
