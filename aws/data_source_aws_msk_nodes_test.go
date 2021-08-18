package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSMskNodesDataSource_Name(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_msk_nodes.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   testAccErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskNodesDataSourceConfigName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "nodes.#", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttr(dataSourceName, "nodes.0.broker_id", "1"),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.0.attached_eni_id", regexp.MustCompile(`^eni-.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.0.client_subnet", regexp.MustCompile(`^subnet-.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.0.endpoints.0", regexp.MustCompile(`^[\w\-\.]+\.kafka\.[\w\-]+\.amazonaws.com$`)),
					resource.TestCheckResourceAttr(dataSourceName, "nodes.1.broker_id", "2"),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.1.attached_eni_id", regexp.MustCompile(`^eni-.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.1.client_subnet", regexp.MustCompile(`^subnet-.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.1.endpoints.0", regexp.MustCompile(`^[\w\-\.]+\.kafka\.[\w\-]+\.amazonaws.com$`)),
					resource.TestCheckResourceAttr(dataSourceName, "nodes.2.broker_id", "3"),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.2.attached_eni_id", regexp.MustCompile(`^eni-.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.2.client_subnet", regexp.MustCompile(`^subnet-.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "nodes.2.endpoints.0", regexp.MustCompile(`^[\w\-\.]+\.kafka\.[\w\-]+\.amazonaws.com$`)),
				),
			},
		},
	})
}
func testAccMskNodesDataSourceConfigName(rName string) string {
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

data "aws_msk_nodes" "test" {
  cluster_arn = aws_msk_cluster.test.arn
}
`, rName))
}
