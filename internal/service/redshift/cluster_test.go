// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, "cluster_nodes.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_nodes.0.public_ip_address"),
					resource.TestCheckResourceAttr(resourceName, "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrDNSName, regexache.MustCompile(fmt.Sprintf("^%s.*\\.redshift\\..*", rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "aqua_configuration_status", "auto"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_track_name", "current"),
					resource.TestCheckResourceAttr(resourceName, "manual_snapshot_retention_period", "-1"),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccRedshiftCluster_aqua(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_aqua(rName, names.AttrEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "aqua_configuration_status", "auto"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccClusterConfig_aqua(rName, "disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "aqua_configuration_status", "auto"),
				),
			},
			{
				Config: testAccClusterConfig_aqua(rName, names.AttrEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "aqua_configuration_status", "auto"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftCluster_withFinalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterTestSnapshotDestroy(ctx, rName),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_finalSnapshot(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccRedshiftCluster_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	keyResourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kmsKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, keyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccRedshiftCluster_enhancedVPCRoutingEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_enhancedVPCRoutingEnabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enhanced_vpc_routing", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccClusterConfig_enhancedVPCRoutingDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enhanced_vpc_routing", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_iamRoles(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_iamRoles(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "2"),
				),
			},
			{
				Config: testAccClusterConfig_updateIAMRoles(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "1"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_publiclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_publiclyAccessible(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
				),
			},

			{
				Config: testAccClusterConfig_publiclyAccessible(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_publiclyAccessible_default(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.RedshiftServiceID),
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.92.0",
					},
				},
				Config: testAccClusterConfig_publiclyAccessible_default(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
				),
			},
			{
				// plan should not empty because the default value has changed and will for an update unless explicitly set
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccClusterConfig_publiclyAccessible_default(rName),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccClusterConfig_publiclyAccessible_default(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_updateNodeCount(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "1"),
				),
			},
			{
				Config: testAccClusterConfig_updateNodeCount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_type", "multi-node"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "dc2.large"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_updateNodeType(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_updateNodeType(rName, "dc2.large"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "node_type", "dc2.large"),
				),
			},
			{
				Config: testAccClusterConfig_updateNodeType(rName, "dc2.8xlarge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "node_type", "dc2.8xlarge"),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccClusterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_masterUsername(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_masterUsername(rName, "foo_test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v1),
					testAccCheckClusterMasterUsername(&v1, "foo_test"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "foo_test"),
				),
			},
			{
				Config: testAccClusterConfig_masterUsername(rName, "new-username"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v2),
					testAccCheckClusterMasterUsername(&v2, "new-username"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "new-username"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_masterUsername_invalid(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_masterUsername(rName, "invalid username"), // no spaces
				ExpectError: regexache.MustCompile(`Error: invalid value for master_username`),
			},
			{
				Config:      testAccClusterConfig_masterUsername(rName, "1invalidusername"), // must start with letter
				ExpectError: regexache.MustCompile(`Error: invalid value for master_username`),
			},
			{
				Config:      testAccClusterConfig_masterUsername(rName, "-invalidusername"), // must start with letter
				ExpectError: regexache.MustCompile(`Error: invalid value for master_username`),
			},
		},
	})
}

func TestAccRedshiftCluster_changeAvailabilityZone(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_updateAvailabilityZone(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, "data.aws_availability_zones.available", "names.0"),
				),
			},
			{
				Config: testAccClusterConfig_updateAvailabilityZone(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, "data.aws_availability_zones.available", "names.1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_changeAvailabilityZoneAndSetAvailabilityZoneRelocation(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_updateAvailabilityZoneAvailabilityZoneRelocationNotSet(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, "data.aws_availability_zones.available", "names.0"),
				),
			},
			{
				Config: testAccClusterConfig_updateAvailabilityZone(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, "data.aws_availability_zones.available", "names.1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_changeAvailabilityZone_availabilityZoneRelocationNotSet(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_updateAvailabilityZoneAvailabilityZoneRelocationNotSet(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, "data.aws_availability_zones.available", "names.0"),
				),
			},
			{
				Config:      testAccClusterConfig_updateAvailabilityZoneAvailabilityZoneRelocationNotSet(rName, 1),
				ExpectError: regexache.MustCompile("cannot change `availability_zone` if `availability_zone_relocation_enabled` is not true"),
			},
		},
	})
}

func TestAccRedshiftCluster_changeEncryption_unsetToFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_encrypted(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.StringExact(acctest.CtFalse)),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_changeEncryption_unsetToTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_encrypted(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_changeEncryption_trueToFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_encrypted(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.StringExact(acctest.CtFalse)),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_changeEncryption_falseToTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
				),
			},
			{
				Config: testAccClusterConfig_encrypted(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.StringExact(acctest.CtTrue)),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_changeEncryption_falseToUnset(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
				),
			},
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.StringExact(acctest.CtTrue)),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_changeEncryption_trueToUnset(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encrypted(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
			},
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_Migrate_encrypted_default(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.RedshiftServiceID),
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.97.0",
					},
				},
				Config: testAccClusterConfig_encrypted_default(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.Bool(false)),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.Bool(false)),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccClusterConfig_encrypted_default(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_Migrate_encrypted_true(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.RedshiftServiceID),
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.97.0",
					},
				},
				Config: testAccClusterConfig_encrypted(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccClusterConfig_encrypted(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_Migrate_encrypted_false(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.RedshiftServiceID),
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.97.0",
					},
				},
				Config: testAccClusterConfig_encrypted(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.Bool(false)),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.Bool(false)),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// Needs a second apply to actually set `encrypted` to `false`
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.97.0",
					},
				},
				Config: testAccClusterConfig_encrypted(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncrypted), knownvalue.Bool(false)),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccClusterConfig_encrypted(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccRedshiftCluster_availabilityZoneRelocation(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_availabilityZoneRelocation(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccClusterConfig_availabilityZoneRelocation(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_availabilityZoneRelocation_publiclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_availabilityZoneRelocationPubliclyAccessible(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_relocation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_restoreFromSnapshot_Identifier(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.restored"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_restoreFromSnapshot_Identifier(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_identifier", "aws_redshift_cluster_snapshot.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccRedshiftCluster_restoreFromSnapshot_ARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.restored"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_restoreFromSnapshot_ARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_arn", "aws_redshift_cluster_snapshot.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					"snapshot_arn",
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccRedshiftCluster_restoreFromSnapshot_ChangeEncryption_trueToFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.restored"
	sourceResourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_restoreFromSnapshot_ChangeEncryption(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(sourceResourceName, names.AttrEncrypted, acctest.CtTrue),

					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccRedshiftCluster_restoreFromSnapshot_ChangeEncryption_falseToTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.restored"
	sourceResourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_restoreFromSnapshot_ChangeEncryption(rName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(sourceResourceName, names.AttrEncrypted, acctest.CtFalse),

					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccRedshiftCluster_manageMasterPassword(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_manageMasterPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "manage_master_password", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "master_password_secret_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"manage_master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccRedshiftCluster_multiAZ(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_multiAZ(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrFinalSnapshotIdentifier,
					"master_password",
					"skip_final_snapshot",
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccClusterConfig_multiAZ(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRedshiftCluster_passwordWriteOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Cluster
	resourceName := "aws_redshift_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.RedshiftServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_passwordWriteOnly(rName, "Mustbe8characters", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
				),
			},
			{
				Config: testAccClusterConfig_passwordWriteOnly(rName, "Mustbe8charactersupdated", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_cluster" {
				continue
			}

			_, err := tfredshift.FindClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterTestSnapshotDestroy(ctx context.Context, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

			_, err := conn.DeleteClusterSnapshot(ctx, &redshift.DeleteClusterSnapshotInput{
				SnapshotIdentifier: aws.String(rName),
			})

			if err != nil {
				return fmt.Errorf("deleting Redshift Cluster Snapshot (%s): %w", rName, err)
			}

			_, err = tfredshift.FindClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, n string, v *awstypes.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		output, err := tfredshift.FindClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterMasterUsername(c *awstypes.Cluster, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if got, want := aws.ToString(c.MasterUsername), value; got != want {
			return fmt.Errorf("Expected cluster's MasterUsername: %q, given: %q", want, got)
		}
		return nil
	}
}

func testAccClusterConfig_updateNodeCount(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  number_of_nodes                     = 2
  skip_final_snapshot                 = true
}
`, rName)
}

func testAccClusterConfig_updateNodeType(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = %[2]q
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  number_of_nodes                     = 2
  skip_final_snapshot                 = true
}
`, rName, nodeType)
}

func testAccClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName)
}

func testAccClusterConfig_aqua(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
  aqua_configuration_status           = %[2]q
  apply_immediately                   = true
}
`, rName, status)
}

func testAccClusterConfig_encrypted(rName string, encrypted bool) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
  publicly_accessible                 = false

  encrypted = %[2]t
}
`, rName, encrypted)
}

func testAccClusterConfig_encrypted_default(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
  publicly_accessible                 = false
}
`, rName)
}

func testAccClusterConfig_finalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = false
  final_snapshot_identifier           = %[1]q
}
`, rName)
}

func testAccClusterConfig_kmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  kms_key_id                          = aws_kms_key.test.arn
  encrypted                           = true
  skip_final_snapshot                 = true
}
`, rName)
}

func testAccClusterConfig_enhancedVPCRoutingEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  enhanced_vpc_routing                = true
  skip_final_snapshot                 = true
}
`, rName)
}

func testAccClusterConfig_enhancedVPCRoutingDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  enhanced_vpc_routing                = false
  skip_final_snapshot                 = true
}
`, rName)
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 7
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 7
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccClusterConfig_publiclyAccessible(rName string, publiclyAccessible bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 3),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  cluster_subnet_group_name           = aws_redshift_subnet_group.test.name
  publicly_accessible                 = %[2]t
  skip_final_snapshot                 = true

  depends_on = [aws_internet_gateway.test]
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName, publiclyAccessible))
}

func testAccClusterConfig_publiclyAccessible_default(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 3),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true # required for v5.92.0
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  cluster_subnet_group_name = aws_redshift_subnet_group.test.name

  depends_on = [aws_internet_gateway.test]
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName))
}

func testAccClusterConfig_iamRoles(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "ec2" {
  name = "%[1]s-ec2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "lambda.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  iam_roles                           = [aws_iam_role.ec2.arn, aws_iam_role.lambda.arn]
  skip_final_snapshot                 = true
}
`, rName)
}

func testAccClusterConfig_updateIAMRoles(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "ec2" {
  name = "%[1]s-ec2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "lambda.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  iam_roles                           = [aws_iam_role.ec2.arn]
  skip_final_snapshot                 = true
}
`, rName)
}

func testAccClusterConfig_updateAvailabilityZone(rName string, regionIndex int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  publicly_accessible                  = false
  availability_zone_relocation_enabled = true
  availability_zone                    = data.aws_availability_zones.available.names[%[2]d]
}
`, rName, regionIndex))
}

func testAccClusterConfig_updateAvailabilityZoneAvailabilityZoneRelocationNotSet(rName string, regionIndex int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  publicly_accessible                  = false
  availability_zone_relocation_enabled = false
  availability_zone                    = data.aws_availability_zones.available.names[%[2]d]
}
`, rName, regionIndex))
}

func testAccClusterConfig_availabilityZoneRelocation(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  number_of_nodes                     = 2
  cluster_type                        = "multi-node"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  publicly_accessible                  = false
  availability_zone_relocation_enabled = %[2]t
}
`, rName, enabled))
}

func testAccClusterConfig_availabilityZoneRelocationPubliclyAccessible(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 3),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  publicly_accessible                  = true
  availability_zone_relocation_enabled = true
  cluster_subnet_group_name            = aws_redshift_subnet_group.test.name

  depends_on = [aws_internet_gateway.test]
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName))
}

func testAccClusterConfig_restoreFromSnapshot_Identifier(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_redshift_cluster_snapshot" "test" {
  cluster_identifier  = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier = %[1]q
}

resource "aws_redshift_cluster" "restored" {
  cluster_identifier  = "%[1]s-restored"
  snapshot_identifier = aws_redshift_cluster_snapshot.test.id
  database_name       = "mydb"
  encrypted           = true
  master_username     = "foo_test"
  master_password     = "Mustbe8characters"
  node_type           = "dc2.large"
  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_restoreFromSnapshot_ARN(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_redshift_cluster_snapshot" "test" {
  cluster_identifier  = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier = %[1]q
}

resource "aws_redshift_cluster" "restored" {
  cluster_identifier  = "%[1]s-restored"
  snapshot_arn        = aws_redshift_cluster_snapshot.test.arn
  database_name       = "mydb"
  encrypted           = true
  master_username     = "foo_test"
  master_password     = "Mustbe8characters"
  node_type           = "dc2.large"
  skip_final_snapshot = true
}
`, rName))
}

func testAccClusterConfig_restoreFromSnapshot_ChangeEncryption(rName string, sourceEncrypted, restoreEncrypted bool) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_encrypted(rName, sourceEncrypted),
		fmt.Sprintf(`
resource "aws_redshift_cluster_snapshot" "test" {
  cluster_identifier  = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier = %[1]q
}

resource "aws_redshift_cluster" "restored" {
  cluster_identifier  = "%[1]s-restored"
  snapshot_identifier = aws_redshift_cluster_snapshot.test.id
  database_name       = "mydb"
  master_username     = "foo_test"
  master_password     = "Mustbe8characters"
  node_type           = "dc2.large"
  skip_final_snapshot = true

  encrypted = %[2]t
}
`, rName, restoreEncrypted))
}

func testAccClusterConfig_manageMasterPassword(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  manage_master_password              = true
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName)
}

func testAccClusterConfig_multiAZ(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "ra3.xlplus"
  number_of_nodes                     = 2
  cluster_type                        = "multi-node"
  automated_snapshot_retention_period = 1
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
  encrypted                           = true
  kms_key_id                          = aws_kms_key.test.arn

  publicly_accessible                  = false
  availability_zone_relocation_enabled = false
  multi_az                             = %[2]t
}
`, rName, enabled))
}

func testAccClusterConfig_passwordWriteOnly(rName, password string, passwordVersion int) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = "foo_test"
  master_password_wo                  = %[2]q
  master_password_wo_version          = %[3]d
  multi_az                            = false
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName, password, passwordVersion)
}

func testAccClusterConfig_masterUsername(rName, username string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  database_name                       = "mydb"
  encrypted                           = true
  master_username                     = %[2]q
  master_password                     = "Mustbe8characters"
  multi_az                            = false
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rName, username)
}
