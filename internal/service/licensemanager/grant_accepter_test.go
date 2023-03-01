package licensemanager_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflicensemanager "github.com/hashicorp/terraform-provider-aws/internal/service/licensemanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccLicenseManagerGrantAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	grantARNKey := "LICENSE_MANAGER_GRANT_ACCEPTER_ARN_BASIC"
	grantARN := os.Getenv(grantARNKey)
	if grantARN == "" {
		t.Skipf("Environment variable %s is not set to true", grantARNKey)
	}
	resourceName := "aws_licensemanager_grant_accepter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGrantAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantAccepterConfig_basic(grantARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGrantAccepterExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "grant_arn"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "ListPurchasedLicenses"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "CheckoutLicense"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "CheckInLicense"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "ExtendConsumptionLicense"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "CreateToken"),
					resource.TestCheckResourceAttrSet(resourceName, "home_region"),
					resource.TestCheckResourceAttrSet(resourceName, "license_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "parent_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "principal"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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

func TestAccLicenseManagerGrantAccepter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	grantARNKey := "LICENSE_MANAGER_GRANT_ACCEPTER_ARN_DISAPPEARS"
	grantARN := os.Getenv(grantARNKey)
	if grantARN == "" {
		t.Skipf("Environment variable %s is not set to true", grantARNKey)
	}
	resourceName := "aws_licensemanager_grant_accepter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGrantAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantAccepterConfig_basic(grantARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantAccepterExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflicensemanager.ResourceGrantAccepter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGrantAccepterExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No License Manager License Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn()

		out, err := tflicensemanager.FindGrantAccepterByGrantARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("GrantAccepter %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGrantAccepterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_licensemanager_grant_accepter" {
				continue
			}

			_, err := tflicensemanager.FindGrantAccepterByGrantARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("License Manager GrantAccepter %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGrantAccepterConfig_basic(grantARN string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_grant_accepter" "test" {
  grant_arn = %[1]q
}
`, grantARN)
}
