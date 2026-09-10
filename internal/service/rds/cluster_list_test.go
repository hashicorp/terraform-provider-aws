// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"testing"

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

func TestAccRDSCluster_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_rds_cluster.test[0]"
	resourceName2 := "aws_rds_cluster.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Cluster/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("rds", "cluster:"+rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("rds", "cluster:"+rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Cluster/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_rds_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_rds_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_rds_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_rds_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccRDSCluster_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_rds_cluster.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Cluster/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("rds", "cluster:"+rName+"-0")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Cluster/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_rds_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_rds_cluster.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAllocatedStorage), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAllowMajorVersionUpgrade), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrApplyImmediately), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("rds", "cluster:"+rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAutoMinorVersionUpgrade), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAvailabilityZones), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("backtrack_window"), knownvalue.Int64Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("backup_retention_period"), knownvalue.Int64Exact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ca_certificate_identifier"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ca_certificate_valid_till"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrClusterIdentifier), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("cluster_identifier_prefix"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("cluster_members"), knownvalue.SetSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("cluster_resource_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("cluster_scalability_type"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("copy_tags_to_snapshot"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("database_insights_mode"), knownvalue.StringExact("standard")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDatabaseName), knownvalue.StringExact("test")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("db_cluster_instance_class"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("db_cluster_parameter_group_name"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("db_instance_parameter_group_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("db_subnet_group_name"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("db_system_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("delete_automated_backups"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDeletionProtection), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDomain), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("domain_iam_role_name"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_global_write_forwarding"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_http_endpoint"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_local_write_forwarding"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enabled_cloudwatch_logs_exports"), knownvalue.SetSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEngine), knownvalue.StringExact("aurora-mysql")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("engine_lifecycle_support"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("engine_mode"), knownvalue.StringExact("provisioned")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEngineVersion), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("engine_version_actual"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrFinalSnapshotIdentifier), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("global_cluster_identifier"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrHostedZoneID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("iam_database_authentication_enabled"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("iam_roles"), knownvalue.SetSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrIOPS), knownvalue.Int64Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrKMSKeyID), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("manage_master_user_password"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("master_password"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("master_password_wo_version"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("master_user_secret"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("master_user_secret_kms_key_id"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("master_username"), knownvalue.StringExact("tfacctest")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("monitoring_interval"), knownvalue.Int64Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("monitoring_role_arn"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("network_type"), knownvalue.StringExact("IPV4")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("performance_insights_enabled"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("performance_insights_kms_key_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("performance_insights_retention_period"), knownvalue.Int64Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPort), knownvalue.Int64Exact(3306)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("preferred_backup_window"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPreferredMaintenanceWindow), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("reader_endpoint"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("replication_source_identifier"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("restore_to_point_in_time"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("s3_import"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("scaling_configuration"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("serverlessv2_scaling_configuration"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("skip_final_snapshot"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("snapshot_identifier"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("source_region"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStorageEncrypted), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStorageType), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("upgrade_rollout_order"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCSecurityGroupIDs), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccRDSCluster_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_rds_cluster.test[0]"
	resourceName2 := "aws_rds_cluster.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Cluster/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("rds", "cluster:"+rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("rds", "cluster:"+rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Cluster/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster.test", identity2.Checks()),
				},
			},
		},
	})
}
