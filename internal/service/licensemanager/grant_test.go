package licensemanager_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

func TestAccLicenseManagerGrant_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"grant": {
			"basic":      testAccLicenseManagerGrant_basic,
			"disappears": testAccLicenseManagerGrant_disappears,
			"name":       testAccLicenseManagerGrant_name,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccLicenseManagerGrant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	principalKey := "LICENSE_MANAGER_GRANT_PRINCIPAL"
	licenseKey := "LICENSE_MANAGER_GRANT_LICENSE_ARN"
	homeRegionKey := "LICENSE_MANAGER_GRANT_HOME_REGION"
	principal := os.Getenv(principalKey)
	if principal == "" {
		t.Skipf("Environment variable %s is not set to true", principalKey)
	}
	licenseArn := os.Getenv(licenseKey)
	if licenseArn == "" {
		t.Skipf("Environment variable %s is not set to true", licenseKey)
	}
	homeRegion := os.Getenv(homeRegionKey)
	if homeRegion == "" {
		t.Skipf("Environment variable %s is not set to true", homeRegionKey)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_grant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantConfig_basic(rName, licenseArn, principal, homeRegion),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGrantExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "license-manager", regexp.MustCompile(`grant:g-.+`)),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "ListPurchasedLicenses"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "CheckoutLicense"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "CheckInLicense"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "ExtendConsumptionLicense"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_operations.*", "CreateToken"),
					resource.TestCheckResourceAttr(resourceName, "home_region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "license_arn", "arn:aws:license-manager::294406891311:license:l-ecbaa94eb71a4830b6d7e49268fecaa0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "parent_arn"),
					resource.TestCheckResourceAttr(resourceName, "principal", "arn:aws:iam::067863992282:root"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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

func testAccLicenseManagerGrant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	principalKey := "LICENSE_MANAGER_GRANT_PRINCIPAL"
	licenseKey := "LICENSE_MANAGER_GRANT_LICENSE_ARN"
	homeRegionKey := "LICENSE_MANAGER_GRANT_HOME_REGION"
	principal := os.Getenv(principalKey)
	if principal == "" {
		t.Skipf("Environment variable %s is not set to true", principalKey)
	}
	licenseArn := os.Getenv(licenseKey)
	if licenseArn == "" {
		t.Skipf("Environment variable %s is not set to true", licenseKey)
	}
	homeRegion := os.Getenv(homeRegionKey)
	if homeRegion == "" {
		t.Skipf("Environment variable %s is not set to true", homeRegionKey)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_grant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantConfig_basic(rName, licenseArn, principal, homeRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflicensemanager.ResourceGrant(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccLicenseManagerGrant_name(t *testing.T) {
	ctx := acctest.Context(t)
	principalKey := "LICENSE_MANAGER_GRANT_PRINCIPAL"
	licenseKey := "LICENSE_MANAGER_GRANT_LICENSE_ARN"
	homeRegionKey := "LICENSE_MANAGER_GRANT_HOME_REGION"
	principal := os.Getenv(principalKey)
	if principal == "" {
		t.Skipf("Environment variable %s is not set to true", principalKey)
	}
	licenseArn := os.Getenv(licenseKey)
	if licenseArn == "" {
		t.Skipf("Environment variable %s is not set to true", licenseKey)
	}
	homeRegion := os.Getenv(homeRegionKey)
	if homeRegion == "" {
		t.Skipf("Environment variable %s is not set to true", homeRegionKey)
	}
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_grant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGrantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGrantConfig_basic(rName1, licenseArn, principal, homeRegion),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGrantExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGrantConfig_basic(rName2, licenseArn, principal, homeRegion),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGrantExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func testAccCheckGrantExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No License Manager License Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn()

		_, err := tflicensemanager.FindGrantByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckGrantDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_licensemanager_grant" {
				continue
			}

			_, err := tflicensemanager.FindGrantByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("License Manager Grant %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGrantConfig_basic(rName string, licenseArn string, principal string, homeRegion string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_grant" "test" {
  name                  = %[1]q
  allowed_operations = [
	"ListPurchasedLicenses",
	"CheckoutLicense",
	"CheckInLicense",
	"ExtendConsumptionLicense",
	"CreateToken"
  ]
  license_arn = %[2]q
  principal = %[3]q
  home_region = %[4]q
}
`, rName, licenseArn, principal, homeRegion)
}
