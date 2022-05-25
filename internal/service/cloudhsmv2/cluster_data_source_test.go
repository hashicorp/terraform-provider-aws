package cloudhsmv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccDataSourceCluster_basic(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.cluster"
	dataSourceName := "data.aws_cloudhsm_v2_cluster.default"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudhsmv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_state", "UNINITIALIZED"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_id", resourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_state", resourceName, "cluster_state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

var testAccClusterDataSourceConfig_basic = acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "cloudhsm_v2_test_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-aws_cloudhsm_v2_cluster-data-source-basic"
  }
}

resource "aws_subnet" "cloudhsm_v2_test_subnets" {
  count                   = 2
  vpc_id                  = aws_vpc.cloudhsm_v2_test_vpc.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = false
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-data-source-basic"
  }
}

resource "aws_cloudhsm_v2_cluster" "cluster" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.cloudhsm_v2_test_subnets[*].id

  tags = {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-data-source-basic-%d"
  }
}

data "aws_cloudhsm_v2_cluster" "default" {
  cluster_id = aws_cloudhsm_v2_cluster.cluster.cluster_id
}
`, sdkacctest.RandInt()))
