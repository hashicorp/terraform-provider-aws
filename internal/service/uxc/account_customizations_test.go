// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package uxc_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfuxc "github.com/hashicorp/terraform-provider-aws/internal/service/uxc"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// All UXC tests run serially — the service is account-scoped and tests would interfere if run in parallel.
func TestAccUXC_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"accountCustomizationsBasic":          testAccAccountCustomizations_basic,
		"accountCustomizationsVisibleRegions": testAccAccountCustomizations_visibleRegions,
		"accountCustomizationsVisibleServices": testAccAccountCustomizations_visibleServices,
		"accountCustomizationsDisappears":     testAccAccountCustomizations_disappears,
		"servicesDataSourceBasic":             testAccServicesDataSource_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountCustomizations_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_uxc_account_customizations.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.UXCServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountCustomizationsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountCustomizationsConfig_basic("pink"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_color", "pink"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountCustomizationsConfig_basic("purple"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_color", "purple"),
				),
			},
		},
	})
}

func testAccAccountCustomizations_visibleRegions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_uxc_account_customizations.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.UXCServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountCustomizationsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountCustomizationsConfig_visibleRegions([]string{"us-east-1", "us-west-2"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "visible_regions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "visible_regions.*", "us-east-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "visible_regions.*", "us-west-2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountCustomizationsConfig_visibleRegions([]string{"eu-west-1"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "visible_regions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "visible_regions.*", "eu-west-1"),
				),
			},
		},
	})
}

func testAccAccountCustomizations_visibleServices(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_uxc_account_customizations.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.UXCServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountCustomizationsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountCustomizationsConfig_visibleServices([]string{"s3", "ec2"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "visible_services.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "visible_services.*", "s3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "visible_services.*", "ec2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountCustomizationsConfig_visibleServices([]string{"s3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "visible_services.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "visible_services.*", "s3"),
				),
			},
		},
	})
}

func testAccAccountCustomizations_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_uxc_account_customizations.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.UXCServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountCustomizationsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountCustomizationsConfig_basic("pink"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfuxc.ResourceAccountCustomizations, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountCustomizationsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).UXCClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_uxc_account_customizations" {
				continue
			}

			output, err := tfuxc.FindAccountCustomizations(ctx, conn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			// Resource has no delete - verify it was reset to defaults.
			if string(output.AccountColor) != "none" && string(output.AccountColor) != "" {
				return errors.New("UXC Account Customizations account_color was not reset to default")
			}
			if len(output.VisibleRegions) > 0 {
				return errors.New("UXC Account Customizations visible_regions was not reset to default")
			}
			if len(output.VisibleServices) > 0 {
				return errors.New("UXC Account Customizations visible_services was not reset to default")
			}
		}

		return nil
	}
}

func testAccCheckAccountCustomizationsExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).UXCClient(ctx)

		_, err := tfuxc.FindAccountCustomizations(ctx, conn)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).UXCClient(ctx)

	_, err := tfuxc.FindAccountCustomizations(ctx, conn)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAccountCustomizationsConfig_basic(color string) string {
	return fmt.Sprintf(`
resource "aws_uxc_account_customizations" "test" {
  account_color = %[1]q
}
`, color)
}

func testAccAccountCustomizationsConfig_visibleRegions(regions []string) string {
	quoted := make([]string, len(regions))
	for i, r := range regions {
		quoted[i] = fmt.Sprintf("%q", r)
	}
	return fmt.Sprintf(`
resource "aws_uxc_account_customizations" "test" {
  visible_regions = [%s]
}
`, strings.Join(quoted, ", "))
}

func testAccAccountCustomizationsConfig_visibleServices(services []string) string {
	quoted := make([]string, len(services))
	for i, s := range services {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf(`
resource "aws_uxc_account_customizations" "test" {
  visible_services = [%s]
}
`, strings.Join(quoted, ", "))
}
