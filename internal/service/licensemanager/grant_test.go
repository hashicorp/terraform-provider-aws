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
			"basic":      testAccGrant_basic,
			"disappears": testAccGrant_disappears,
			"name":       testAccGrant_name,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccGrant_basic(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "home_region", homeRegion),
					resource.TestCheckResourceAttr(resourceName, "license_arn", licenseArn),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "parent_arn"),
					resource.TestCheckResourceAttr(resourceName, "principal", principal),
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

func testAccGrant_disappears(t *testing.T) {
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

func testAccGrant_name(t *testing.T) {
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

		out, err := tflicensemanager.FindGrantByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Bucket %q does not exist", rs.Primary.ID)
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
  name = %[1]q
  allowed_operations = [
    "ListPurchasedLicenses",
    "CheckoutLicense",
    "CheckInLicense",
    "ExtendConsumptionLicense",
    "CreateToken"
  ]
  license_arn = %[2]q
  principal   = %[3]q
  home_region = %[4]q
}
`, rName, licenseArn, principal, homeRegion)
}
