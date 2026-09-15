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

func TestAccRDSClusterInstance_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_rds_cluster_instance.test[0]"
	resourceName2 := "aws_rds_cluster_instance.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ClusterInstance/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("rds", "db:"+rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("rds", "db:"+rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ClusterInstance/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster_instance.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_rds_cluster_instance.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_rds_cluster_instance.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster_instance.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_rds_cluster_instance.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_rds_cluster_instance.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_rds_cluster_instance.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ClusterInstance/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("rds", "db:"+rName+"-0")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ClusterInstance/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster_instance.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_rds_cluster_instance.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_rds_cluster_instance.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrApplyImmediately), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("rds", "db:"+rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAutoMinorVersionUpgrade), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAvailabilityZone), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ca_cert_identifier"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrClusterIdentifier), knownvalue.StringExact(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("copy_tags_to_snapshot"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("custom_iam_instance_profile"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("db_parameter_group_name"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("db_subnet_group_name"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("dbi_resource_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEngine), knownvalue.StringExact("aurora-mysql")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEngineVersion), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("engine_version_actual"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrForceDestroy), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrIdentifier), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("identifier_prefix"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("instance_class"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrKMSKeyID), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("monitoring_interval"), knownvalue.Int64Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("monitoring_role_arn"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("network_type"), knownvalue.StringExact("IPV4")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("performance_insights_enabled"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("performance_insights_kms_key_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("performance_insights_retention_period"), knownvalue.Int64Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPort), knownvalue.Int64Exact(3306)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("preferred_backup_window"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPreferredMaintenanceWindow), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("promotion_tier"), knownvalue.Int64Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPubliclyAccessible), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStorageEncrypted), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("warning_event_categories"), knownvalue.SetSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("writer"), knownvalue.Bool(true)),
					}),
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_rds_cluster_instance.test[0]"
	resourceName2 := "aws_rds_cluster_instance.test[1]"
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
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ClusterInstance/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("rds", "db:"+rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("rds", "db:"+rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ClusterInstance/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster_instance.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_rds_cluster_instance.test", identity2.Checks()),
				},
			},
		},
	})
}
