// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccODBExascaleDBStorageVault_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exascaleDBStorageVault odbtypes.ExascaleDbStorageVault
	rName := acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)
	resourceName := "aws_odb_exascale_db_storage_vault.test"
	availabilityZoneID := testAccExascaleDBStorageVaultAvailabilityZoneIDs[acctest.Region()]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExascaleDBStorageVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExascaleDBStorageVaultConfig_basic(rName, availabilityZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExascaleDBStorageVaultExists(ctx, t, resourceName, &exascaleDBStorageVault),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+$`)),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_id", availabilityZoneID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "high_capacity_database_storage_total_size_in_gbs", "300"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBExascaleDBStorageVault_allArguments(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exascaleDBStorageVault odbtypes.ExascaleDbStorageVault
	rName := acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)
	resourceName := "aws_odb_exascale_db_storage_vault.test"
	availabilityZoneID := testAccExascaleDBStorageVaultAvailabilityZoneIDs[acctest.Region()]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExascaleDBStorageVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExascaleDBStorageVaultConfig_allArguments(rName, availabilityZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExascaleDBStorageVaultExists(ctx, t, resourceName, &exascaleDBStorageVault),
					resource.TestCheckResourceAttr(resourceName, "additional_flash_cache_in_percent", "34"),
					resource.TestCheckResourceAttr(resourceName, "autoscale_limit_in_gbs", "600"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, "data.aws_availability_zone.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_id", availabilityZoneID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName+" description"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "high_capacity_database_storage_total_size_in_gbs", "300"),
					resource.TestCheckResourceAttr(resourceName, "is_autoscale_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccODBExascaleDBStorageVault_update(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exascaleDBStorageVault odbtypes.ExascaleDbStorageVault
	rName := acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)
	resourceName := "aws_odb_exascale_db_storage_vault.test"
	availabilityZoneID := testAccExascaleDBStorageVaultAvailabilityZoneIDs[acctest.Region()]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExascaleDBStorageVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExascaleDBStorageVaultConfig_updateBefore(rName, availabilityZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExascaleDBStorageVaultExists(ctx, t, resourceName, &exascaleDBStorageVault),
					resource.TestCheckResourceAttr(resourceName, "additional_flash_cache_in_percent", "34"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName+" initial description"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "high_capacity_database_storage_total_size_in_gbs", "300"),
					resource.TestCheckResourceAttr(resourceName, "is_autoscale_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccExascaleDBStorageVaultConfig_updateAfter(rName, availabilityZoneID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExascaleDBStorageVaultExists(ctx, t, resourceName, &exascaleDBStorageVault),
					resource.TestCheckResourceAttr(resourceName, "additional_flash_cache_in_percent", "51"),
					resource.TestCheckResourceAttr(resourceName, "autoscale_limit_in_gbs", "1200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName+" updated description"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "high_capacity_database_storage_total_size_in_gbs", "400"),
					resource.TestCheckResourceAttr(resourceName, "is_autoscale_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccExascaleDBStorageVaultConfig_updateAfterOmitAutoscaleLimit(rName, availabilityZoneID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExascaleDBStorageVaultExists(ctx, t, resourceName, &exascaleDBStorageVault),
					resource.TestCheckResourceAttr(resourceName, "autoscale_limit_in_gbs", "1200"),
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

func TestAccODBExascaleDBStorageVault_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)
	resourceName := "aws_odb_exascale_db_storage_vault.test"
	availabilityZoneID := testAccExascaleDBStorageVaultAvailabilityZoneIDs[acctest.Region()]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExascaleDBStorageVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExascaleDBStorageVaultConfig_basic(rName, availabilityZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfodb.ResourceExascaleDBStorageVault, resourceName),
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

func testAccCheckExascaleDBStorageVaultDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_exascale_db_storage_vault" {
				continue
			}

			_, err := tfodb.FindExascaleDBStorageVaultByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameExascaleDBStorageVault, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckExascaleDBStorageVaultExists(ctx context.Context, t *testing.T, name string, exascaleDBStorageVault *odbtypes.ExascaleDbStorageVault) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameExascaleDBStorageVault, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameExascaleDBStorageVault, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
		out, err := tfodb.FindExascaleDBStorageVaultByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameExascaleDBStorageVault, rs.Primary.ID, err)
		}

		*exascaleDBStorageVault = *out

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)

	_, err := conn.ListExascaleDbStorageVaults(ctx, &odb.ListExascaleDbStorageVaultsInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const testAccExascaleDBStorageVaultDisplayNamePrefix = "ofake"

func testAccRandomExascaleDBStorageVaultDisplayName(t *testing.T) string {
	return acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)
}

var testAccExascaleDBStorageVaultAvailabilityZoneIDs = map[string]string{
	endpoints.EuWest1RegionID: "euw1-az1",
	endpoints.UsEast1RegionID: "use1-az2",
}

func testAccExascaleDBStorageVaultConfig_basic(rName, availabilityZoneID string) string {
	return fmt.Sprintf(`
resource "aws_odb_exascale_db_storage_vault" "test" {
  availability_zone_id                             = %[2]q
  display_name                                     = %[1]q
  high_capacity_database_storage_total_size_in_gbs = 300
}
`, rName, availabilityZoneID)
}

func testAccExascaleDBStorageVaultConfig_allArguments(rName, availabilityZoneID string) string {
	return fmt.Sprintf(`
data "aws_availability_zone" "test" {
  zone_id = %[2]q
}

resource "aws_odb_exascale_db_storage_vault" "test" {
  additional_flash_cache_in_percent                = 34
  autoscale_limit_in_gbs                           = 600
  availability_zone                                = data.aws_availability_zone.test.name
  availability_zone_id                             = %[2]q
  description                                      = "%[1]s description"
  display_name                                     = %[1]q
  high_capacity_database_storage_total_size_in_gbs = 300
  is_autoscale_enabled                             = true
  time_zone                                        = "UTC"

  tags = {
    Name = %[1]q
  }
}
`, rName, availabilityZoneID)
}

func testAccExascaleDBStorageVaultConfig_updateBefore(rName, availabilityZoneID string) string {
	return fmt.Sprintf(`
resource "aws_odb_exascale_db_storage_vault" "test" {
  additional_flash_cache_in_percent                = 34
  availability_zone_id                             = %[2]q
  description                                      = "%[1]s initial description"
  display_name                                     = %[1]q
  high_capacity_database_storage_total_size_in_gbs = 300
  is_autoscale_enabled                             = false
}
`, rName, availabilityZoneID)
}

func testAccExascaleDBStorageVaultConfig_updateAfter(rName, availabilityZoneID string) string {
	return fmt.Sprintf(`
resource "aws_odb_exascale_db_storage_vault" "test" {
  additional_flash_cache_in_percent                = 51
  autoscale_limit_in_gbs                           = 1200
  availability_zone_id                             = %[2]q
  description                                      = "%[1]s updated description"
  display_name                                     = "%[1]s-updated"
  high_capacity_database_storage_total_size_in_gbs = 400
  is_autoscale_enabled                             = true
}
`, rName, availabilityZoneID)
}

func testAccExascaleDBStorageVaultConfig_updateAfterOmitAutoscaleLimit(rName, availabilityZoneID string) string {
	return fmt.Sprintf(`
resource "aws_odb_exascale_db_storage_vault" "test" {
  additional_flash_cache_in_percent                = 51
  availability_zone_id                             = %[2]q
  description                                      = "%[1]s updated description"
  display_name                                     = "%[1]s-updated"
  high_capacity_database_storage_total_size_in_gbs = 400
  is_autoscale_enabled                             = true
}
`, rName, availabilityZoneID)
}
