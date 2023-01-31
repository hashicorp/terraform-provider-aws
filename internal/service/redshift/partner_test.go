package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftPartner_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_partner.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartnerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPartnerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartnerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "partner_name", "Datacoral"),
					resource.TestCheckResourceAttrPair(resourceName, "database_name", "aws_redshift_cluster.test", "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", "aws_redshift_cluster.test", "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"account_id", "cluster_identifier"},
			},
		},
	})
}

func TestAccRedshiftPartner_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_partner.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartnerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPartnerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartnerExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourcePartner(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftPartner_disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_partner.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPartnerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPartnerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPartnerExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceCluster(), "aws_redshift_cluster.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPartnerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_partner" {
				continue
			}
			_, err := tfredshift.FindPartnerById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckPartnerExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Partner ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn()

		_, err := tfredshift.FindPartnerById(ctx, conn, rs.Primary.ID)

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
