// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSDirectorySettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_directory_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DSEndpointID)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectorySettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectorySettingsConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectorySettingsExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("directory_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("TLS_1_0")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrValue), knownvalue.StringExact("Disable")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "directory_id", names.AttrName),
				ImportStateVerifyIdentifierAttribute: "directory_id",
				ImportStateVerify:                    true,
			},
		},
	})
}

func TestAccDSDirectorySettings_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_directory_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DSEndpointID)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectorySettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectorySettingsConfig_value(rName, domainName, "Disable"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrValue), knownvalue.StringExact("Disable")),
				},
			},
			{
				Config: testAccDirectorySettingsConfig_value(rName, domainName, "Enable"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrValue), knownvalue.StringExact("Enable")),
				},
			},
		},
	})
}

// TestAccDSDirectorySettings_disappears_Directory verifies that Settings is
// correctly detected as gone when its backing directory disappears out of band.
// directory_settings implements framework.WithNoOpDelete, so the parent
// directory is disappeared instead.
func TestAccDSDirectorySettings_disappears_Directory(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_directory_settings.test"
	directoryResourceName := "aws_directory_service_directory.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DSEndpointID)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectorySettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectorySettingsConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectorySettingsExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfds.ResourceDirectory(), directoryResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckDirectorySettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_directory_settings" {
				continue
			}

			directoryID := rs.Primary.Attributes["directory_id"]
			name := rs.Primary.Attributes[names.AttrName]

			_, err := tfds.FindDirectorySettingByTwoPartKey(ctx, conn, directoryID, name)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DS, create.ErrActionCheckingDestroyed, tfds.ResNameDirectorySettingsExported, directoryID, err)
			}

			return create.Error(names.DS, create.ErrActionCheckingDestroyed, tfds.ResNameDirectorySettingsExported, directoryID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDirectorySettingsExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameDirectorySettingsExported, name, errors.New("not found"))
		}

		directoryID := rs.Primary.Attributes["directory_id"]
		settingName := rs.Primary.Attributes[names.AttrName]
		if directoryID == "" || settingName == "" {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameDirectorySettingsExported, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		_, err := tfds.FindDirectorySettingByTwoPartKey(ctx, conn, directoryID, settingName)
		return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameDirectorySettingsExported, directoryID, err)
	}
}

func testAccDirectorySettingsConfig_basic(rName, domain string) string {
	return testAccDirectorySettingsConfig_value(rName, domain, "Disable")
}

func testAccDirectorySettingsConfig_value(rName, domain, value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory_settings" "test" {
  directory_id = aws_directory_service_directory.test.id
  name         = "TLS_1_0"
  value        = %[2]q
}
`, domain, value),
	)
}
