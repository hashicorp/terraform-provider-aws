// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeAuthPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var authpolicy vpclattice.GetAuthPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_auth_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthPolicyExists(ctx, resourceName, &authpolicy),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"Action":"*"`)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_identifier", "aws_vpclattice_service.test", names.AttrARN),
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

func TestAccVPCLatticeAuthPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var authpolicy vpclattice.GetAuthPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_auth_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthPolicyExists(ctx, resourceName, &authpolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceAuthPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAuthPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_auth_policy" {
				continue
			}

			_, err := tfvpclattice.FindAuthPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPCLattice Auth Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAuthPolicyExists(ctx context.Context, n string, v *vpclattice.GetAuthPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		output, err := tfvpclattice.FindAuthPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAuthPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_vpclattice_service" "test" {
  name               = %[1]q
  auth_type          = "AWS_IAM"
  custom_domain_name = "example.com"
}

resource "aws_vpclattice_auth_policy" "test" {
  resource_identifier = aws_vpclattice_service.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "*"
      Effect    = "Allow"
      Principal = "*"
      Resource  = "*"
      Condition = {
        StringNotEqualsIgnoreCase = {
          "aws:PrincipalType" = "anonymous"
        }
      }
    }]
  })
}
`, rName)
}
