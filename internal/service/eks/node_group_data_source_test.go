package eks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEKSNodeGroupDataSource_basic(t *testing.T) {
	var nodeGroup eks.Nodegroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_node_group.test"
	resourceName := "aws_eks_node_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeGroupConfig_dataSourceName(rName),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccNodeGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNodeGroupExists(resourceName, &nodeGroup),
					resource.TestCheckResourceAttrPair(resourceName, "ami_type", dataSourceResourceName, "ami_type"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", dataSourceResourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(resourceName, "disk_size", dataSourceResourceName, "disk_size"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "instance_types.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_type", dataSourceResourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(resourceName, "labels.%", dataSourceResourceName, "labels.%"),
					resource.TestCheckResourceAttrPair(resourceName, "node_group_name", dataSourceResourceName, "node_group_name"),
					resource.TestCheckResourceAttrPair(resourceName, "node_role_arn", dataSourceResourceName, "node_role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "release_version", dataSourceResourceName, "release_version"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "remote_access.#", "0"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "resources.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "resources", dataSourceResourceName, "resources"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "scaling_config", dataSourceResourceName, "scaling_config"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceResourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", dataSourceResourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids", dataSourceResourceName, "subnet_ids"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "taints.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dataSourceResourceName, "version"),
				),
			},
		},
	})
}

func testAccNodeGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNodeGroupConfig_dataSourceName(rName), fmt.Sprintf(`
data "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
}
`, rName))
}
