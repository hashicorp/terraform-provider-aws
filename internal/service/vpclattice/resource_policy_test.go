// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy vpclattice.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"vpc-lattice:CreateServiceNetworkVpcAssociation","vpc-lattice:CreateServiceNetworkServiceAssociation","vpc-lattice:GetServiceNetwork"`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_vpclattice_service_network.test", names.AttrARN),
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

func TestAccVPCLatticeResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy vpclattice.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					acctest.CheckSDKResourceDisappears(ctx, t, tfvpclattice.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_resource_policy" {
				continue
			}

			_, err := tfvpclattice.FindResourcePolicyByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPCLattice Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, n string, v *vpclattice.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).VPCLatticeClient(ctx)

		output, err := tfvpclattice.FindResourcePolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpclattice_resource_policy" "test" {
  resource_arn = aws_vpclattice_service_network.test.arn

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Sid    = "test-pol-principals-6"
      Effect = "Allow"
      Principal = {
        "AWS" = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "vpc-lattice:CreateServiceNetworkVpcAssociation",
        "vpc-lattice:CreateServiceNetworkServiceAssociation",
        "vpc-lattice:GetServiceNetwork"
      ]
      Resource = aws_vpclattice_service_network.test.arn
    }]
  })
}
`, rName)
}
