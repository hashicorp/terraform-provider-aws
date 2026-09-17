// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package uxc_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/uxc/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfuxc "github.com/hashicorp/terraform-provider-aws/internal/service/uxc"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccUXC_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AccountCustomizations": {
			acctest.CtBasic:        testAccAccountCustomizations_basic,
			"accountColor":         testAccAccountCustomizations_accountColor,
			"visibleRegions":       testAccAccountCustomizations_visibleRegions,
			"visibleRegionsEmpty":  testAccAccountCustomizations_visibleRegionsEmpty,
			"visibleRegionsNil":    testAccAccountCustomizations_visibleRegionsNil,
			"visibleServices":      testAccAccountCustomizations_visibleServices,
			"visibleServicesEmpty": testAccAccountCustomizations_visibleServicesEmpty,
			"visibleServicesNil":   testAccAccountCustomizations_visibleServicesNil,
			acctest.CtDisappears:   testAccAccountCustomizations_disappears,
			"migrate":              testAccAccountCustomizations_Migrate_basic,
			"Identity":             testAccUXCAccountCustomizations_identitySerial,
		},
		"ServicesDataSource": {
			acctest.CtBasic: testAccServicesDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
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
				Config: testAccAccountCustomizationsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("account_color"), tfknownvalue.StringExact(awstypes.AccountColorNone)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "account_color",
			},
		},
	})
}

func testAccAccountCustomizations_accountColor(t *testing.T) {
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
				Config: testAccAccountCustomizationsConfig_accountColor("pink"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_color", "pink"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "account_color",
			},
			{
				Config: testAccAccountCustomizationsConfig_accountColor("purple"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_color", "purple"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "account_color",
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
				Config: testAccAccountCustomizationsConfig_visibleRegions([]string{
					"us-east-1", //lintignore:AWSAT003
					"us-west-2", //lintignore:AWSAT003
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("us-east-1"), //lintignore:AWSAT003
						knownvalue.StringExact("us-west-2"), //lintignore:AWSAT003
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "account_color",
			},
			{
				Config: testAccAccountCustomizationsConfig_visibleRegions([]string{
					"eu-west-1", //lintignore:AWSAT003
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("eu-west-1"), //lintignore:AWSAT003
					})),
				},
			},
		},
	})
}

func testAccAccountCustomizations_visibleRegionsEmpty(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"onCreate": testAccAccountCustomizations_visibleRegionsEmpty_onCreate,
		"onUpdate": testAccAccountCustomizations_visibleRegionsEmpty_onUpdate,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountCustomizations_visibleRegionsEmpty_onCreate(t *testing.T) {
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
				Config: testAccAccountCustomizationsConfig_visibleRegionsEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "account_color",
			},
		},
	})
}

func testAccAccountCustomizations_visibleRegionsEmpty_onUpdate(t *testing.T) {
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
				Config: testAccAccountCustomizationsConfig_visibleRegions([]string{"us-east-1"}), //lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("us-east-1"), //lintignore:AWSAT003
					})),
				},
			},
			{
				// Explicitly set an empty set — must not cause a permanent diff.
				Config: testAccAccountCustomizationsConfig_visibleRegionsEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				// Confirm no diff on subsequent plan.
				Config: testAccAccountCustomizationsConfig_visibleRegionsEmpty(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccAccountCustomizations_visibleRegionsNil(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"onUpdate": testAccAccountCustomizations_visibleRegionsNil_onUpdate,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountCustomizations_visibleRegionsNil_onUpdate(t *testing.T) {
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
				Config: testAccAccountCustomizationsConfig_visibleRegions([]string{"us-east-1"}), //lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("us-east-1"), //lintignore:AWSAT003
					})),
				},
			},
			{
				// Explicitly set to nil — must not cause a permanent diff.
				Config: testAccAccountCustomizationsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				// Confirm no diff on subsequent plan.
				Config: testAccAccountCustomizationsConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("s3"),
						knownvalue.StringExact("ec2"),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "account_color",
			},
			{
				Config: testAccAccountCustomizationsConfig_visibleServices([]string{"s3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("s3"),
					})),
				},
			},
		},
	})
}

func testAccAccountCustomizations_visibleServicesEmpty(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"onCreate": testAccAccountCustomizations_visibleServicesEmpty_onCreate,
		"onUpdate": testAccAccountCustomizations_visibleServicesEmpty_onUpdate,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountCustomizations_visibleServicesEmpty_onCreate(t *testing.T) {
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
				Config: testAccAccountCustomizationsConfig_visibleServicesEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "account_color",
			},
		},
	})
}

func testAccAccountCustomizations_visibleServicesEmpty_onUpdate(t *testing.T) {
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
				Config: testAccAccountCustomizationsConfig_visibleServices([]string{"s3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("s3"),
					})),
				},
			},
			{
				// Explicitly set an empty set — must not cause a permanent diff.
				Config: testAccAccountCustomizationsConfig_visibleServicesEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				// Confirm no diff on subsequent plan.
				Config: testAccAccountCustomizationsConfig_visibleServicesEmpty(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccAccountCustomizations_visibleServicesNil(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"onUpdate": testAccAccountCustomizations_visibleServicesNil_onUpdate,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountCustomizations_visibleServicesNil_onUpdate(t *testing.T) {
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
				Config: testAccAccountCustomizationsConfig_visibleServices([]string{"s3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("s3"),
					})),
				},
			},
			{
				// Explicitly set nil — must not cause a permanent diff.
				Config: testAccAccountCustomizationsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				// Confirm no diff on subsequent plan.
				Config: testAccAccountCustomizationsConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
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
				Config: testAccAccountCustomizationsConfig_accountColor("pink"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfuxc.ResourceAccountCustomizations, resourceName), // nosemgrep:disappears-expect-resource-action
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccAccountCustomizations_Migrate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_uxc_account_customizations.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.UXCServiceID),
		CheckDestroy: testAccCheckAccountCustomizationsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.49.0",
					},
				},
				Config: testAccAccountCustomizationsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("account_color"), tfknownvalue.StringExact(awstypes.AccountColorNone)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountCustomizationsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountCustomizationsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("account_color"), tfknownvalue.StringExact(awstypes.AccountColorNone)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_regions"), knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("visible_services"), knownvalue.SetSizeExact(0)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
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

func testAccAccountCustomizationsConfig_basic() string {
	return `
resource "aws_uxc_account_customizations" "test" {}
`
}

func testAccAccountCustomizationsConfig_accountColor(color string) string {
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

func testAccAccountCustomizationsConfig_visibleRegionsEmpty() string {
	return `
resource "aws_uxc_account_customizations" "test" {
  visible_regions = []
}
`
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

func testAccAccountCustomizationsConfig_visibleServicesEmpty() string {
	return `
resource "aws_uxc_account_customizations" "test" {
  visible_services = []
}
`
}
