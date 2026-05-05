// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSRadiusSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DIRECTORY_SERVICE_RADIUS_SERVER"
	radiusServer := os.Getenv(key)
	if radiusServer == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v awstypes.RadiusSettings
	resourceName := "aws_directory_service_region.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRadiusSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRadiusSettingsConfig_basic(rName, domainName, radiusServer),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRadiusSettingsExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_protocol", "PAP"),
					resource.TestCheckResourceAttr(resourceName, "display_label", "test"),
					resource.TestCheckResourceAttr(resourceName, "radius_port", "1812"),
					resource.TestCheckResourceAttr(resourceName, "radius_retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "radius_servers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "radius_servers.*", radiusServer),
					resource.TestCheckResourceAttr(resourceName, "radius_timeout", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "shared_secret"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"shared_secret",
				},
			},
		},
	})
}

func TestAccDSRadiusSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DIRECTORY_SERVICE_RADIUS_SERVER"
	radiusServer := os.Getenv(key)
	if radiusServer == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v awstypes.RadiusSettings
	resourceName := "aws_directory_service_region.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRadiusSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRadiusSettingsConfig_basic(rName, domainName, radiusServer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRadiusSettingsExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfds.ResourceRadiusSettings(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRadiusSettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_radius_settings" {
				continue
			}

			_, err := tfds.FindRadiusSettingsByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service Directory %s RADIUS Settings still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRadiusSettingsExists(ctx context.Context, t *testing.T, n string, v *awstypes.RadiusSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		output, err := tfds.FindRadiusSettingsByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRadiusSettingsConfig_basic(rName, domain, radiusServer string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_radius_settings" "test" {
  directory_id = aws_directory_service_directory.test.id

  authentication_protocol = "PAP"
  display_label           = "test"
  radius_port             = 1812
  radius_retries          = 3
  radius_servers          = [%[2]q]
  radius_timeout          = 30
  shared_secret           = "avoid-plaintext-passwords"
}
`, domain, radiusServer))
}
