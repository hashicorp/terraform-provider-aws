package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enable_default_standards"},
			},
		},
	})
}

func testAccAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccount_enableDefaultStandardsFalse(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_enableDefaultStandardsFalse,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", "false"),
				),
			},
		},
	})
}

func testAccCheckAccountExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub Account ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn()

		_, err := tfsecurityhub.FindStandardsSubscriptions(ctx, conn, &securityhub.GetEnabledStandardsInput{})

		return err
	}
}

func testAccCheckAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_account" {
				continue
			}

			_, err := tfsecurityhub.FindStandardsSubscriptions(ctx, conn, &securityhub.GetEnabledStandardsInput{})

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccAccountConfig_basic = `
resource "aws_securityhub_account" "test" {}
`

const testAccAccountConfig_enableDefaultStandardsFalse = `
resource "aws_securityhub_account" "test" {
  enable_default_standards = false
}
`
