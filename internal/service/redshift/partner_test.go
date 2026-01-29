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

func TestAccRedshiftPartner_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_partner.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartnerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPartnerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartnerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "partner_name", "Datacoral"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDatabaseName, "aws_redshift_cluster.test", names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrID),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrAccountID, names.AttrClusterIdentifier},
			},
		},
	})
}

func TestAccRedshiftPartner_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_partner.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartnerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPartnerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartnerExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfredshift.ResourcePartner(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftPartner_disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_partner.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartnerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPartnerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartnerExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfredshift.ResourceCluster(), "aws_redshift_cluster.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPartnerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_partner" {
				continue
			}
			_, err := tfredshift.FindPartnerByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes[names.AttrClusterIdentifier], rs.Primary.Attributes[names.AttrDatabaseName], rs.Primary.Attributes["partner_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Partner %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPartnerExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		_, err := tfredshift.FindPartnerByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes[names.AttrClusterIdentifier], rs.Primary.Attributes[names.AttrDatabaseName], rs.Primary.Attributes["partner_name"])

		return err
	}
}

func testAccPartnerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), `
data "aws_caller_identity" "current" {}

resource "aws_redshift_partner" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  account_id         = data.aws_caller_identity.current.account_id
  database_name      = aws_redshift_cluster.test.database_name
  partner_name       = "Datacoral"
}
`)
}
