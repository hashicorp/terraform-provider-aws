// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2DedicatedIPAssignment_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccSESV2DedicatedIPAssignment_basic,
		acctest.CtDisappears: testAccSESV2DedicatedIPAssignment_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSESV2DedicatedIPAssignment_basic(t *testing.T) { // nosemgrep:ci.sesv2-in-func-name
	ctx := acctest.Context(t)
	if os.Getenv("SES_DEDICATED_IP") == "" {
		t.Skip("Environment variable SES_DEDICATED_IP is not set")
	}

	ip := os.Getenv("SES_DEDICATED_IP")
	poolName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_assignment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPAssignmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPAssignmentConfig_basic(ip, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPAssignmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip", ip),
					resource.TestCheckResourceAttr(resourceName, "destination_pool_name", poolName),
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

func testAccSESV2DedicatedIPAssignment_disappears(t *testing.T) { // nosemgrep:ci.sesv2-in-func-name
	ctx := acctest.Context(t)
	if os.Getenv("SES_DEDICATED_IP") == "" {
		t.Skip("Environment variable SES_DEDICATED_IP is not set")
	}

	ip := os.Getenv("SES_DEDICATED_IP")
	poolName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_assignment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPAssignmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPAssignmentConfig_basic(ip, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPAssignmentExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceDedicatedIPAssignment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDedicatedIPAssignmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_dedicated_ip_assignment" {
				continue
			}

			_, err := tfsesv2.FindDedicatedIPByTwoPartKey(ctx, conn, rs.Primary.Attributes["ip"], rs.Primary.Attributes["destination_pool_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SESv2 Dedicated IP Assignment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDedicatedIPAssignmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindDedicatedIPByTwoPartKey(ctx, conn, rs.Primary.Attributes["ip"], rs.Primary.Attributes["destination_pool_name"])

		return err
	}
}

func testAccDedicatedIPAssignmentConfig_basic(ip, poolName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_dedicated_ip_pool" "test" {
  pool_name = %[2]q
}

resource "aws_sesv2_dedicated_ip_assignment" "test" {
  ip                    = %[1]q
  destination_pool_name = aws_sesv2_dedicated_ip_pool.test.pool_name
}
`, ip, poolName)
}
