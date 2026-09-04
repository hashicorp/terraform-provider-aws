// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccODBExascaleDBStorageVault_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName1 := "aws_odb_exascale_db_storage_vault.test[0]"
	resourceName2 := "aws_odb_exascale_db_storage_vault.test[1]"
	rName := acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		CheckDestroy:             testAccCheckExascaleDBStorageVaultDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ExascaleDBStorageVault/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+`))),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+`))),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ExascaleDBStorageVault/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_odb_exascale_db_storage_vault.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exascale_db_storage_vault.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_odb_exascale_db_storage_vault.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_odb_exascale_db_storage_vault.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exascale_db_storage_vault.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_odb_exascale_db_storage_vault.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccODBExascaleDBStorageVault_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName1 := "aws_odb_exascale_db_storage_vault.test[0]"
	resourceName2 := "aws_odb_exascale_db_storage_vault.test[1]"
	rName := acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)
	availabilityZoneID := testAccExascaleDBStorageVaultAvailabilityZoneIDs[acctest.Region()]

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		CheckDestroy:             testAccCheckExascaleDBStorageVaultDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ExascaleDBStorageVault/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+`))),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+`))),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ExascaleDBStorageVault/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_odb_exascale_db_storage_vault.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exascale_db_storage_vault.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_odb_exascale_db_storage_vault.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("additional_flash_cache_in_percent"), knownvalue.Int32Exact(34)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("autoscale_limit_in_gbs"), knownvalue.Int32Exact(600)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAvailabilityZone), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("availability_zone_id"), knownvalue.StringExact(availabilityZoneID)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(rName+"-0 description")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("high_capacity_database_storage_total_size_in_gbs"), knownvalue.Int32Exact(300)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("is_autoscale_enabled"), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("time_zone"), knownvalue.StringExact("UTC")),
					}),

					tfquerycheck.ExpectIdentityFunc("aws_odb_exascale_db_storage_vault.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exascale_db_storage_vault.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					querycheck.ExpectResourceKnownValues("aws_odb_exascale_db_storage_vault.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("additional_flash_cache_in_percent"), knownvalue.Int32Exact(34)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("autoscale_limit_in_gbs"), knownvalue.Int32Exact(600)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAvailabilityZone), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("availability_zone_id"), knownvalue.StringExact(availabilityZoneID)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(rName+"-1 description")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(rName+"-1")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("high_capacity_database_storage_total_size_in_gbs"), knownvalue.Int32Exact(300)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("is_autoscale_enabled"), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("time_zone"), knownvalue.StringExact("UTC")),
					}),
				},
			},
		},
	})
}

func TestAccODBExascaleDBStorageVault_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName1 := "aws_odb_exascale_db_storage_vault.test[0]"
	resourceName2 := "aws_odb_exascale_db_storage_vault.test[1]"
	rName := acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ExascaleDBStorageVault/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+`))),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("odb", regexache.MustCompile(`exascale-db-storage-vault/xsvault_.+`))),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ExascaleDBStorageVault/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_odb_exascale_db_storage_vault.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_odb_exascale_db_storage_vault.test", identity2.Checks()),
				},
			},
		},
	})
}
