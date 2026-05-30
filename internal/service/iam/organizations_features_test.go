// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMOrganizationsFeatures_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccOrganizationsFeatures_basic,
		"update":        testAccOrganizationsFeatures_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccOrganizationsFeatures_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_organizations_features.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "iam.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationsFeaturesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationsFeaturesConfig_basic([]string{"RootCredentialsManagement", "RootSessions"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationsFeaturesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_features"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_features"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("RootCredentialsManagement"),
						knownvalue.StringExact("RootSessions"),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func testAccOrganizationsFeatures_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_organizations_features.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "iam.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationsFeaturesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationsFeaturesConfig_basic([]string{"RootCredentialsManagement"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationsFeaturesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_features"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_features"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("RootCredentialsManagement"),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccOrganizationsFeaturesConfig_basic([]string{"RootSessions"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationsFeaturesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_features"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_features"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("RootSessions"),
					})),
				},
			},
		},
	})
}

func testAccCheckOrganizationsFeaturesDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_organizations_features" {
				continue
			}

			_, err := tfiam.FindOrganizationsFeatures(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Organizations Features %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationsFeaturesExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		_, err := tfiam.FindOrganizationsFeatures(ctx, conn)

		return err
	}
}

func testAccOrganizationsFeaturesConfig_basic(features []string) string {
	return fmt.Sprintf(`
resource "aws_iam_organizations_features" "test" {
  enabled_features = [%[1]s]
}
`, fmt.Sprintf(`"%s"`, strings.Join(features, `", "`)))
}
