package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSMskBrokerNodesDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_msk_broker_nodes.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   testAccErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskBrokerNodesDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_info_list.#", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttr(dataSourceName, "node_info_list.0.broker_id", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "node_info_list.1.broker_id", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "node_info_list.2.broker_id", "3"),
				),
			},
		},
	})
}
func testAccMskBrokerNodesDataSourceConfig(rName string) string {
	return composeConfig(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.2.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.example_sg.id]
  }

  tags = {
    foo = "bar"
  }
}

data "aws_msk_broker_nodes" "test" {
  cluster_arn = aws_msk_cluster.test.arn
}
`, rName))
}
