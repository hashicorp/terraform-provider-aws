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
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.Region()]
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_basic/"),
				ConfigVariables: testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, availabilityZoneID, publicKey, 2),
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
				ConfigVariables: testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, availabilityZoneID, publicKey, 2),
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
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.Region()]
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	variables := testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, availabilityZoneID, publicKey, 2)
	variables[acctest.CtResourceTags] = config.MapVariable(map[string]config.Variable{
		acctest.CtKey1: config.StringVariable(acctest.CtValue1),
	})

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
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
					testAccExaDBVMClusterListKnownValues(identity1.Checks(), rName+"-0", "ofake0"+hostnameSuffix, publicKey),
					tfquerycheck.ExpectIdentityFunc("aws_odb_exadb_vm_cluster.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					testAccExaDBVMClusterListKnownValues(identity2.Checks(), rName+"-1", "ofake1"+hostnameSuffix, publicKey),
				},
			},
		},
	})
}

func TestAccODBExaDBVMCluster_List_storageVaultFilter(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName1 := "aws_odb_exadb_vm_cluster.test[0]"
	resourceName2 := "aws_odb_exadb_vm_cluster.test[1]"
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostnameSuffix := acctest.RandStringFromCharSet(t, 5, acctest.CharSetAlphaNum)
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.Region()]
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_storage_vault_filter/"),
				ConfigVariables: testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, availabilityZoneID, publicKey, 2),
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ExaDBVMCluster/list_storage_vault_filter/"),
				ConfigVariables: testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, availabilityZoneID, publicKey, 2),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("aws_odb_exadb_vm_cluster.test", 1),
					tfquerycheck.ExpectIdentityFunc("aws_odb_exadb_vm_cluster.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectNoIdentityFunc("aws_odb_exadb_vm_cluster.test", identity2.Checks()),
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
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.AlternateRegion()]
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	variables := testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, availabilityZoneID, publicKey, 2)
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

func testAccExaDBVMClusterListConfigVariables(rName, hostnameSuffix, availabilityZoneID, publicKey string, resourceCount int) config.Variables {
	return config.Variables{
		acctest.CtRName:        config.StringVariable(rName),
		"availability_zone_id": config.StringVariable(availabilityZoneID),
		"hostname_suffix":      config.StringVariable(hostnameSuffix),
		"resource_count":       config.IntegerVariable(resourceCount),
		"ssh_public_key":       config.StringVariable(publicKey),
	}
}

func testAccExaDBVMClusterListKnownValues(identityChecks func() map[string]knownvalue.Check, displayName, hostname, publicKey string) querycheck.QueryResultCheck {
	return querycheck.ExpectResourceKnownValues("aws_odb_exadb_vm_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identityChecks), []querycheck.KnownValueCheck{
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("odb", regexache.MustCompile(`exadb-vm-cluster/.+`))),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrClusterName), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCreatedAt), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("data_collection_options"), knownvalue.ListSizeExact(1)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("data_collection_options").AtSliceIndex(0).AtMapKey("is_diagnostics_events_enabled"), knownvalue.Bool(true)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("data_collection_options").AtSliceIndex(0).AtMapKey("is_health_monitoring_enabled"), knownvalue.Bool(false)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("data_collection_options").AtSliceIndex(0).AtMapKey("is_incident_logs_enabled"), knownvalue.Bool(true)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(displayName)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDomain), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("enabled_ecpu_count"), knownvalue.Int32Exact(testAccExaDBVMClusterEnabledECPUCount)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("exascale_db_storage_vault_arn"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("exascale_db_storage_vault_id"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("gi_version"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("grid_image_id"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("grid_image_type"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("hostname"), knownvalue.StringExact(hostname)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("iam_roles"), knownvalue.Null()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("iorm_config_cache"), knownvalue.Null()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("last_update_history_entry_id"), knownvalue.Null()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("license_model"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("listener_port"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("memory_size_in_gbs"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("node_count"), knownvalue.Int32Exact(testAccExaDBVMClusterNodeCount)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("ocid"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("oci_resource_anchor_name"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("oci_url"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("odb_network_arn"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("odb_network_id"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("percent_progress"), knownvalue.Null()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("scan_dns_name"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("scan_dns_record_id"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("scan_ip_ids"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("scan_listener_port_tcp"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("scan_listener_port_tcp_ssl"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("shape"), knownvalue.StringExact(testAccExaDBVMClusterShapeCanonical)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("shape_attribute"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("snapshot_file_system_storage"), knownvalue.ListSizeExact(1)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("snapshot_file_system_storage").AtSliceIndex(0).AtMapKey("total_size_in_gbs"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("ssh_public_keys"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact(publicKey)})),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("AVAILABLE")),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatusReason), knownvalue.Null()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("system_version"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
			acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
		})),
		tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
			acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
		})),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("time_zone"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("total_ecpu_count"), knownvalue.Int32Exact(testAccExaDBVMClusterTotalECPUCount)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("total_file_system_storage"), knownvalue.ListSizeExact(1)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("total_file_system_storage").AtSliceIndex(0).AtMapKey("total_size_in_gbs"), knownvalue.NotNull()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("vip_ids"), knownvalue.Null()),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("vm_file_system_storage"), knownvalue.ListSizeExact(1)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("vm_file_system_storage").AtSliceIndex(0).AtMapKey("total_size_in_gbs"), knownvalue.Int32Exact(testAccExaDBVMClusterVMFileSystemSizeInGBs)),
		tfquerycheck.KnownValueCheck(tfjsonpath.New("vm_file_system_storage_total_size_in_gbs"), knownvalue.Int32Exact(testAccExaDBVMClusterVMFileSystemSizeInGBs)),
	})
}
