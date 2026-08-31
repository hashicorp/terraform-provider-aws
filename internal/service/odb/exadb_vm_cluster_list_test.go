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

func TestAccODBExaDBVMCluster_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName1 := "aws_odb_exadb_vm_cluster.test[0]"
	resourceName2 := "aws_odb_exadb_vm_cluster.test[1]"
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostnameSuffix := acctest.RandStringFromCharSet(t, 5, acctest.CharSetAlphaNum)
	gridImageID := acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterGridImageIDEnvVar)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_basic/"),
				ConfigVariables: testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, gridImageID, publicKey, 2),
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exadb-vm-cluster/.+`))),
					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exadb-vm-cluster/.+`))),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_basic/"),
				ConfigVariables: testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, gridImageID, publicKey, 2),
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_odb_exadb_vm_cluster.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectIdentityFunc("aws_odb_exadb_vm_cluster.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccODBExaDBVMCluster_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName1 := "aws_odb_exadb_vm_cluster.test[0]"
	resourceName2 := "aws_odb_exadb_vm_cluster.test[1]"
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostnameSuffix := acctest.RandStringFromCharSet(t, 5, acctest.CharSetAlphaNum)
	gridImageID := acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterGridImageIDEnvVar)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	variables := testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, gridImageID, publicKey, 2)
	variables[acctest.CtResourceTags] = config.MapVariable(map[string]config.Variable{
		acctest.CtKey1: config.StringVariable(acctest.CtValue1),
	})

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_include_resource/"),
				ConfigVariables: variables,
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exadb-vm-cluster/.+`))),
					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exadb-vm-cluster/.+`))),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_include_resource/"),
				ConfigVariables: variables,
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_odb_exadb_vm_cluster.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					testAccExaDBVMClusterListKnownValues(identity1.Checks(), rName+"-0", "ofake0"+hostnameSuffix, gridImageID),
					tfquerycheck.ExpectIdentityFunc("aws_odb_exadb_vm_cluster.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					testAccExaDBVMClusterListKnownValues(identity2.Checks(), rName+"-1", "ofake1"+hostnameSuffix, gridImageID),
				},
			},
		},
	})
}

func TestAccODBExaDBVMCluster_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName1 := "aws_odb_exadb_vm_cluster.test[0]"
	resourceName2 := "aws_odb_exadb_vm_cluster.test[1]"
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostnameSuffix := acctest.RandStringFromCharSet(t, 5, acctest.CharSetAlphaNum)
	gridImageID := testAccExaDBVMClusterGridImageIDForRegion(t, acctest.AlternateRegion())
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	variables := testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, gridImageID, publicKey, 2)
	variables[names.AttrRegion] = config.StringVariable(acctest.AlternateRegion())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_region_override/"),
				ConfigVariables: variables,
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("odb", regexache.MustCompile(`exadb-vm-cluster/.+`))),
					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("odb", regexache.MustCompile(`exadb-vm-cluster/.+`))),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_region_override/"),
				ConfigVariables: variables,
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_odb_exadb_vm_cluster.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_odb_exadb_vm_cluster.test", identity2.Checks()),
				},
			},
		},
	})
}

func testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, gridImageID, publicKey string, resourceCount int) config.Variables {
	return config.Variables{
		acctest.CtRName:   config.StringVariable(rName),
		"grid_image_id":   config.StringVariable(gridImageID),
		"hostname_suffix": config.StringVariable(hostnameSuffix),
		"resource_count":  config.IntegerVariable(resourceCount),
		"ssh_public_key":  config.StringVariable(publicKey),
	}
}

func testAccExaDBVMClusterListKnownValues(identityChecks func() map[string]knownvalue.Check, displayName, hostname, gridImageID string) querycheck.QueryResultCheck {
	return querycheck.ExpectResourceKnownValues("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identityChecks), []querycheck.KnownValueCheck{
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exadb-vm-cluster/.+`))),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(displayName)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("enabled_ecpu_count"), knownvalue.Int32Exact(testAccExaDBVMClusterEnabledECPUCount)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("exascale_db_storage_vault_id"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("grid_image_id"), knownvalue.StringExact(gridImageID)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("hostname"), knownvalue.StringExact(hostname)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("node_count"), knownvalue.Int32Exact(testAccExaDBVMClusterNodeCount)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("odb_network_id"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("shape"), knownvalue.StringExact(testAccExaDBVMClusterShape)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("AVAILABLE")),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
			acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
		})),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
			acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
		})),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("total_ecpu_count"), knownvalue.Int32Exact(testAccExaDBVMClusterTotalECPUCount)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("vm_file_system_storage_total_size_in_gbs"), knownvalue.Int32Exact(testAccExaDBVMClusterVMFileSystemSizeInGBs)),
	})
}

func testAccExaDBVMClusterGridImageIDForRegion(t *testing.T, region string) string {
	t.Helper()

	switch region {
	case endpoints.UsEast1RegionID:
		return acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterGridImageIDEnvVar)
	case endpoints.EuWest1RegionID:
		return acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterAlternateGridImageIDEnvVar)
	default:
		t.Fatalf("unsupported ExaDB VM Cluster acceptance test region: %s", region)
		return ""
	}
}
