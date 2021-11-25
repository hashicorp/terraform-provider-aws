package eks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(eks.EndpointsID, testAccErrorCheckSkipEKS)

}

func TestAccEKSClusterRegistration_basic(t *testing.T) {
	var cluster eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster_registration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, eks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterRegistrationBaseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccClusterRegistrationBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccNodeGroupBaseConfig(rName), fmt.Sprintf(`
	resource "aws_eks_node_group" "test" {
	  cluster_name    = aws_eks_cluster.test.name
	  node_group_name = %[1]q
	  node_role_arn   = aws_iam_role.node.arn
	  subnet_ids      = aws_subnet.test[*].id
	
	  scaling_config {
		desired_size = 1
		max_size     = 1
		min_size     = 1
	  }
	
	  depends_on = [
		aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy,
		aws_iam_role_policy_attachment.node-AmazonEKS_CNI_Policy,
		aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly,
	  ]
	}
	`, rName))
}

func testAccClusterRegistrationBaseIAMConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "cluster_registration" {
	name = "%[1]s-role"
	
	assume_role_policy = jsonencode({
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "SSMAccess",
				"Effect": "Allow",
				"Principal": {
					"Service": [
						"ssm.amazonaws.com"
					]
				},
				"Action": "sts:AssumeRole"
			}
		]
	})
}

data aws_iam_policy_document policy_document {
	statement {
		actions = [
			"ssmmessages:CreateControlChannel"
		]

		resources = [
			"arn:aws:eks:*:*:cluster/*"
		] 
	}

	statement {
		actions = [
			"ssmmessages:CreateDataChannel",
			"ssmmessages:OpenDataChannel",
			"ssmmessages:OpenControlChannel"		]

		resources = [
			"*"
		] 
	}
}

resource "aws_iam_policy" "agent_policy" {
	name   = "agent_policy"
	path   = "/"
	policy = data.aws_iam_policy_document.policy_document.json
  }

resource "aws_iam_role_policy_attachment" "test-attach" {
	role       = aws_iam_role.cluster_registration.name
	policy_arn = aws_iam_policy.agent_policy.arn
  }
  
`, rName)
}

func testAccClusterRegistrationBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterRegistrationBaseIAMConfig(rName),
		fmt.Sprintf(`
resource "aws_eks_cluster_registration" "test" {
  name     = %[1]q

  connector_config {
	provider    = "OTHER"
	role_arm    = aws_iam_role.cluster_registration.arn
  }
}
`, rName))
}
