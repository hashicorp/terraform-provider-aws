package eks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEKSNodeGroupsDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_node_groups.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeGroupsDataSourceConfig_namesBasic(rName),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccNodeGroupsDataSourceConfig_names(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(dataSourceResourceName, "names.#", "2"),
				),
			},
		},
	})
}

func testAccNodeGroupsDataSourceConfig_names(rName string) string {
	return acctest.ConfigCompose(testAccNodeGroupsDataSourceConfig_namesBasic(rName), `
data "aws_eks_node_groups" "test" {
  cluster_name = aws_eks_cluster.test.name

  depends_on = [aws_eks_node_group.test_a, aws_eks_node_group.test_b]
}
`)
}

func testAccNodeGroupsDataSourceConfig_namesBasic(rName string) string {
	return acctest.ConfigCompose(testAccNodeGroupBaseConfig(rName), fmt.Sprintf(`
resource "aws_eks_node_group" "test_a" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = "%[1]s-test-a"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }

  depends_on = [
    "aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy",
    "aws_iam_role_policy_attachment.node-AmazonEKS_CNI_Policy",
    "aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly",
  ]
}

resource "aws_eks_node_group" "test_b" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = "%[1]s-test-b"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }

  depends_on = [
    "aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy",
    "aws_iam_role_policy_attachment.node-AmazonEKS_CNI_Policy",
    "aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly",
  ]
}
`, rName))
}
