// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftUsageLimit_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_usage_limit.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_basic(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "feature_type", "concurrency-scaling"),
					resource.TestCheckResourceAttr(resourceName, "limit_type", "time"),
					resource.TestCheckResourceAttr(resourceName, "amount", "60"),
					resource.TestCheckResourceAttr(resourceName, "breach_action", "log"),
					resource.TestCheckResourceAttr(resourceName, "period", "monthly"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
					testAccCheckUsageLimitExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "feature_type", "concurrency-scaling"),
					resource.TestCheckResourceAttr(resourceName, "limit_type", "time"),
					resource.TestCheckResourceAttr(resourceName, "amount", "120"),
					resource.TestCheckResourceAttr(resourceName, "breach_action", "log"),
					resource.TestCheckResourceAttr(resourceName, "period", "monthly"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccRedshiftUsageLimit_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_usage_limit.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageLimitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_basic(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfredshift.ResourceUsageLimit(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUsageLimitDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_usage_limit" {
				continue
			}
			_, err := tfredshift.FindUsageLimitByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Usage Limit %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUsageLimitExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot Copy Grant ID (UsageLimitName) is not set")
		}

		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		_, err := tfredshift.FindUsageLimitByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccUsageLimitConfig_basic(rName string, amount int) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_usage_limit" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  feature_type       = "concurrency-scaling"
  limit_type         = "time"
  amount             = %[1]d
}
`, amount))
}
