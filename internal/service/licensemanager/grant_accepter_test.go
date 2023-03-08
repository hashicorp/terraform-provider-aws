package licensemanager_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/licensemanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflicensemanager "github.com/hashicorp/terraform-provider-aws/internal/service/licensemanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccGrantAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	licenseKey := "LICENSE_MANAGER_GRANT_LICENSE_ARN"
	licenseARN := os.Getenv(licenseKey)
	if licenseARN == "" {
		t.Skipf("Environment variable %s is not set", licenseKey)
	}
	resourceName := "aws_licensemanager_grant_accepter.test"
	resourceGrantName := "aws_licensemanager_grant.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGrantAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantAccepterConfig_basic(licenseARN, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGrantAccepterExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "grant_arn", resourceGrantName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "allowed_operations.0"),
					resource.TestCheckResourceAttrPair(resourceName, "home_region", resourceGrantName, "home_region"),
					resource.TestCheckResourceAttr(resourceName, "license_arn", licenseARN),
					resource.TestCheckResourceAttrPair(resourceName, "name", resourceGrantName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_arn", resourceGrantName, "parent_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", resourceGrantName, "principal"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				Config:            testAccGrantAccepterConfig_basic(licenseARN, rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGrantAccepter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	licenseKey := "LICENSE_MANAGER_GRANT_LICENSE_ARN"
	licenseARN := os.Getenv(licenseKey)
	if licenseARN == "" {
		t.Skipf("Environment variable %s is not set to true", licenseKey)
	}
	resourceName := "aws_licensemanager_grant_accepter.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGrantAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantAccepterConfig_basic(licenseARN, rName),
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

func testAccGrantAccepterConfig_basic(licenseARN string, rName string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_licensemanager_received_license" "test" {
  provider    = awsalternate
  license_arn = %[1]q
}

locals {
  allowed_operations = [for i in data.aws_licensemanager_received_license.test.received_metadata[0].allowed_operations : i if i != "CreateGrant"]
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_licensemanager_grant" "test" {
  provider = awsalternate

  name               = %[2]q
  allowed_operations = local.allowed_operations
  license_arn        = data.aws_licensemanager_received_license.test.license_arn
  principal          = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
}

resource "aws_licensemanager_grant_accepter" "test" {
  grant_arn = aws_licensemanager_grant.test.arn
}
`, licenseARN, rName)
}
