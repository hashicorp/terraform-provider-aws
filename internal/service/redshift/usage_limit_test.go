// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftUsageLimit_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_usage_limit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageLimitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_basic(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "feature_type", "concurrency-scaling"),
					resource.TestCheckResourceAttr(resourceName, "limit_type", "time"),
					resource.TestCheckResourceAttr(resourceName, "amount", "60"),
					resource.TestCheckResourceAttr(resourceName, "breach_action", "log"),
					resource.TestCheckResourceAttr(resourceName, "period", "monthly"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
					resource.TestCheckResourceAttr(resourceName, "feature_type", "concurrency-scaling"),
					resource.TestCheckResourceAttr(resourceName, "limit_type", "time"),
					resource.TestCheckResourceAttr(resourceName, "amount", "120"),
					resource.TestCheckResourceAttr(resourceName, "breach_action", "log"),
					resource.TestCheckResourceAttr(resourceName, "period", "monthly"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRedshiftUsageLimit_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_usage_limit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageLimitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsageLimitConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccUsageLimitConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRedshiftUsageLimit_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_usage_limit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsageLimitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_basic(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceUsageLimit(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUsageLimitDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_usage_limit" {
				continue
			}
			_, err := tfredshift.FindUsageLimitByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckUsageLimitExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot Copy Grant ID (UsageLimitName) is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

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

func testAccUsageLimitConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_usage_limit" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  feature_type       = "concurrency-scaling"
  limit_type         = "time"
  amount             = 60

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccUsageLimitConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_usage_limit" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  feature_type       = "concurrency-scaling"
  limit_type         = "time"
  amount             = 60

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
