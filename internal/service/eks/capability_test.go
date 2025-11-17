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

func TestAccEKSCapability_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var capability types.Capability
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_capability.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, resourceName, &capability),
					resource.TestCheckResourceAttr(resourceName, "capability_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "type", "KRO"),
					resource.TestCheckResourceAttr(resourceName, "delete_propagation_policy", "RETAIN"),
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

func TestAccEKSCapability_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var capability types.Capability
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_capability.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, resourceName, &capability),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfeks.ResourceCapability(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSCapability_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var capability1, capability2, capability3 types.Capability
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_capability.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapabilityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapabilityConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, resourceName, &capability1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapabilityConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, resourceName, &capability2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCapabilityConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapabilityExists(ctx, resourceName, &capability3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckCapabilityExists(ctx context.Context, n string, v *types.Capability) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		clusterName, capabilityName, err := tfeks.CapabilityParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)
		output, err := tfeks.FindCapabilityByTwoPartKey(ctx, conn, clusterName, capabilityName)
		if err != nil {
			return err
		}

		*v = *output
		return nil
	}
}

func testAccCheckCapabilityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eks_capability" {
				continue
			}

			clusterName, capabilityName, err := tfeks.CapabilityParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfeks.FindCapabilityByTwoPartKey(ctx, conn, clusterName, capabilityName)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EKS Capability %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCapabilityConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.cluster.arn

  access_config {
    authentication_mode                         = "API"
    bootstrap_cluster_creator_admin_permissions = true
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy]
}

resource "aws_iam_role" "capability" {
  name = "%[1]s-capability"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "capabilities.eks.amazonaws.com"
      }
      Action = [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }]
  })
}

resource "aws_iam_role_policy_attachment" "capability" {
  role       = aws_iam_role.capability.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AdministratorAccess"
}
`, rName))
}

func testAccCapabilityConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "KRO"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName))
}

func testAccCapabilityConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "KRO"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName, tagKey1, tagValue1))
}

func testAccCapabilityConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccCapabilityConfig_base(rName), fmt.Sprintf(`
resource "aws_eks_capability" "test" {
  cluster_name              = aws_eks_cluster.test.name
  capability_name           = %[1]q
  type                      = "KRO"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.capability]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
