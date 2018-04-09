package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceCloudHsm2Cluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudHsm2ClusterDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudhsm_v2_cluster.default", "cluster_state", "UNINITIALIZED"),
				),
			},
		},
	})
}

var testAccCheckCloudHsm2ClusterDataSourceConfig = fmt.Sprintf(`
variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "cloudhsm2_test_vpc" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "terraform-testacc-aws_cloudhsm_v2_cluster-data-source-basic"
  }
}

resource "aws_subnet" "cloudhsm2_test_subnets" {
  count                   = 2
  vpc_id                  = "${aws_vpc.cloudhsm2_test_vpc.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = false
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-data-source-basic"
  }
}

resource "aws_cloudhsm_v2_cluster" "cluster" {
  hsm_type = "hsm1.medium"  
  subnet_ids = ["${aws_subnet.cloudhsm2_test_subnets.*.id}"]
  tags {
    Name = "tf-acc-aws_cloudhsm_v2_cluster-data-source-basic-%d"
  }
}

data "aws_cloudhsm_v2_cluster" "default" {
  cluster_id = "${aws_cloudhsm_v2_cluster.cluster.cluster_id}"
}
`, acctest.RandInt())
