// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessUsageLimit_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_usage_limit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageLimitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_basic(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_redshiftserverless_workgroup.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "amount", "60"),
					resource.TestCheckResourceAttr(resourceName, "usage_type", "serverless-compute"),
					resource.TestCheckResourceAttr(resourceName, "breach_action", "log"),
					resource.TestCheckResourceAttr(resourceName, "period", "monthly"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsageLimitConfig_basic(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_redshiftserverless_workgroup.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "amount", "120"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessUsageLimit_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshiftserverless_usage_limit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageLimitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_basic(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceUsageLimit(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUsageLimitDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_usage_limit" {
				continue
			}
			_, err := tfredshiftserverless.FindUsageLimitByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Serverless Usage Limit %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUsageLimitExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Serverless Usage Limit is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		_, err := tfredshiftserverless.FindUsageLimitByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccUsageLimitConfig_basic(rName string, amount int) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftserverless_usage_limit" "test" {
  resource_arn = aws_redshiftserverless_workgroup.test.arn
  usage_type   = "serverless-compute"
  amount       = %[2]d
}
`, rName, amount)
}
