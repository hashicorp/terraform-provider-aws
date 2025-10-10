// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud9_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloud9/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloud9 "github.com/hashicorp/terraform-provider-aws/internal/service/cloud9"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloud9EnvironmentMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.EnvironmentMember

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Cloud9EndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Cloud9ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentMembershipConfig_basic(rName, "read-only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentMemberExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrPermissions, "read-only"),
					resource.TestCheckResourceAttrPair(resourceName, "user_arn", "aws_iam_user.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", "aws_cloud9_environment_ec2.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEnvironmentMembershipConfig_basic(rName, "read-write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentMemberExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrPermissions, "read-write"),
					resource.TestCheckResourceAttrPair(resourceName, "user_arn", "aws_iam_user.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", "aws_cloud9_environment_ec2.test", names.AttrID),
				),
			},
		},
	})
}

func TestAccCloud9EnvironmentMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.EnvironmentMember

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Cloud9EndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Cloud9ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentMembershipConfig_basic(rName, "read-only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentMemberExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloud9.ResourceEnvironmentMembership(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloud9EnvironmentMembership_Disappears_env(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.EnvironmentMember

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Cloud9EndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Cloud9ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentMembershipConfig_basic(rName, "read-only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentMemberExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloud9.ResourceEnvironmentEC2(), "aws_cloud9_environment_ec2.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEnvironmentMemberExists(ctx context.Context, n string, v *types.EnvironmentMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Client(ctx)

		output, err := tfcloud9.FindEnvironmentMembershipByTwoPartKey(ctx, conn, rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["user_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEnvironmentMemberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloud9_environment_membership" {
				continue
			}

			_, err := tfcloud9.FindEnvironmentMembershipByTwoPartKey(ctx, conn, rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["user_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cloud9 Environment Membership %s still exists.", rs.Primary.ID)
		}
		return nil
	}
}

func testAccEnvironmentMembershipConfig_basic(rName, permissions string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2Config_basic(rName), fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_cloud9_environment_membership" "test" {
  environment_id = aws_cloud9_environment_ec2.test.id
  permissions    = %[2]q
  user_arn       = aws_iam_user.test.arn
}
`, rName, permissions))
}
