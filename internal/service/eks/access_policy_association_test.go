// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSAccessPolicyAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var associatedaccesspolicy types.AssociatedAccessPolicy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eks_access_policy_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyAssociationExists(ctx, t, resourceName, &associatedaccesspolicy),
					resource.TestCheckResourceAttrSet(resourceName, "associated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrClusterName),
					resource.TestCheckResourceAttrSet(resourceName, "modified_at"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "principal_arn"),
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

func TestAccEKSAccessPolicyAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var associatedaccesspolicy types.AssociatedAccessPolicy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eks_access_policy_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyAssociationExists(ctx, t, resourceName, &associatedaccesspolicy),
					acctest.CheckSDKResourceDisappears(ctx, t, tfeks.ResourceAccessPolicyAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSAccessPolicyAssociation_Disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var associatedaccesspolicy types.AssociatedAccessPolicy
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eks_access_policy_association.test"
	clusterResourceName := "aws_eks_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyAssociationExists(ctx, t, resourceName, &associatedaccesspolicy),
					acctest.CheckSDKResourceDisappears(ctx, t, tfeks.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessPolicyAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EKSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_access_policy_association" {
				continue
			}

			_, err := tfeks.FindAccessPolicyAssociationByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrClusterName], rs.Primary.Attributes["principal_arn"], rs.Primary.Attributes["policy_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EKS Access Policy Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessPolicyAssociationExists(ctx context.Context, t *testing.T, n string, v *types.AssociatedAccessPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EKSClient(ctx)

		output, err := tfeks.FindAccessPolicyAssociationByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrClusterName], rs.Primary.Attributes["principal_arn"], rs.Primary.Attributes["policy_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccessPolicyAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_service_principal" "eks" {
  service_name = "eks"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "${data.aws_service_principal.eks.name}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  access_config {
    authentication_mode = "API"
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccAccessPolicyAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccessPolicyAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_eks_access_entry" "test" {
  cluster_name  = aws_eks_cluster.test.name
  principal_arn = aws_iam_user.test.arn
  depends_on    = [aws_eks_cluster.test]
}

resource "aws_eks_access_policy_association" "test" {
  cluster_name  = aws_eks_cluster.test.name
  principal_arn = aws_iam_user.test.arn
  policy_arn    = "arn:${data.aws_partition.current.partition}:eks::aws:cluster-access-policy/AmazonEKSViewPolicy"

  access_scope {
    type = "cluster"
  }
  depends_on = [aws_eks_cluster.test, aws_eks_access_entry.test]
}
`, rName))
}
