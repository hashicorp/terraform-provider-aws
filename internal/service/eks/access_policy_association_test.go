// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSAccessPolicyAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var associatedaccesspolicy types.AssociatedAccessPolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_access_policy_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyAssociationExists(ctx, resourceName, &associatedaccesspolicy),
					resource.TestCheckResourceAttrSet(resourceName, "associated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_name"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_access_policy_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyAssociationExists(ctx, resourceName, &associatedaccesspolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfeks.ResourceAccessPolicyAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessPolicyAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_access_policy_association" {
				continue
			}

			clusterName, principal_arn, policy_arn, err := tfeks.AssociatePolicyParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}
			_, err = tfeks.FindAccessPolicyByID(ctx, conn, clusterName, principal_arn, policy_arn)

			if tfresource.NotFound(err) {
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

func testAccCheckAccessPolicyAssociationExists(ctx context.Context, name string, associatedaccesspolicy *types.AssociatedAccessPolicy) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		clusterName, principal_arn, policy_arn, err := tfeks.AssociatePolicyParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		output, err := tfeks.FindAccessPolicyByID(ctx, conn, clusterName, principal_arn, policy_arn)
		if err != nil {
			return err
		}

		*associatedaccesspolicy = *output

		return nil
	}
}

func testAccAccessPolicyAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.${data.aws_partition.current.dns_suffix}"
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
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy"

  access_scope {
    type = "cluster"
  }
  depends_on = [aws_eks_cluster.test, aws_eks_access_entry.test]
}
`, rName))
}
