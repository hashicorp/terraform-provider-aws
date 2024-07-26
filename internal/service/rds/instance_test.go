// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, acctest.Ct10),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "backup_target", names.AttrRegion),
					resource.TestCheckResourceAttrSet(resourceName, "backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert_identifier"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "db_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "dedicated_log_volume", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.InstanceEngineMySQL),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "license_model", "general-public-license"),
					resource.TestCheckResourceAttr(resourceName, "listener_endpoint.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, "option_group_name", regexache.MustCompile(`^default:mysql-\d`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrParameterGroupName, regexache.MustCompile(`^default\.mysql\d`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "3306"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replicas.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "available"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, "tfacctest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"manage_master_user_password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_identifierPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_identifierPrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrIdentifier, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_identifierGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_identifierGenerated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSInstance_engineLifecycleSupport_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_engineLifecycleSupport_disabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.InstanceEngineMySQL),
					resource.TestCheckResourceAttr(resourceName, "engine_lifecycle_support", "open-source-rds-extended-support-disabled"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"manage_master_user_password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_Versions_onlyMajor(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_majorVersionOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.InstanceEngineMySQL),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "8.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrEngineVersion,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_kmsKeyID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"delete_automated_backups",
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_customIAMInstanceProfile(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_customIAMInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "custom_iam_instance_profile"),
				),
			},
		},
	})
}

func TestAccRDSInstance_DBSubnetGroupName_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	dbSubnetGroupResourceName2 := "aws_db_subnet_group.test2"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_DBSubnetGroupName_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
			{
				Config: testAccInstanceConfig_DBSubnetGroupName_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName2, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", fmt.Sprintf("%s-2", rName)),
				),
			},
		},
	})
}

func TestAccRDSInstance_networkType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkType(rName, "IPV4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
				),
			},
			{
				Config: testAccInstanceConfig_networkType(rName, "DUAL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "network_type", "DUAL"),
				),
			},
		},
	})
}

func TestAccRDSInstance_optionGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_optionGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "option_group_name", rName),
				),
			},
		},
	})
}

func TestAccRDSInstance_iamAuth(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_iamAuth(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_Versions_allowMajor(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1 rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Versions_allowMajor(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrAllowMajorVersionUpgrade,
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_Versions_allowMajor(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSInstance_db2(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	// Requires an IBM Db2 License set as environmental variable.
	// Licensing pre-requisite: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/db2-licensing.html.
	customerID := acctest.SkipIfEnvVarNotSet(t, "RDS_DB2_CUSTOMER_ID")
	siteID := acctest.SkipIfEnvVarNotSet(t, "RDS_DB2_SITE_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_db2engine(rName, customerID, siteID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_DBSubnetGroupName_ramShared(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_DBSubnetGroupName_ramShared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccRDSInstance_DBSubnetGroupName_vpcSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_DBSubnetGroupName_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccRDSInstance_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_deletionProtection(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
			{
				Config: testAccInstanceConfig_deletionProtection(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSInstance_FinalSnapshotIdentifier_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// testAccCheckInstanceDestroyWithFinalSnapshot verifies a database snapshot is
		// created, and subsequently deletes it
		CheckDestroy: testAccCheckInstanceDestroyWithFinalSnapshot(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_finalSnapshotID(rName1, rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
				),
			},
			// Test updating just final_snapshot_identifier.
			// https://github.com/hashicorp/terraform-provider-aws/issues/26280
			{
				Config: testAccInstanceConfig_finalSnapshotID(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func TestAccRDSInstance_FinalSnapshotIdentifier_skipFinalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroyWithoutFinalSnapshot(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_FinalSnapshotID_skipFinalSnapshot(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func TestAccRDSInstance_isAlreadyBeingDeleted(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mariaDB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
			{
				PreConfig: func() {
					// Get Database Instance into deleting state
					conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)
					input := &rds.DeleteDBInstanceInput{
						DBInstanceIdentifier: aws.String(rName),
						SkipFinalSnapshot:    aws.Bool(true),
					}
					_, err := conn.DeleteDBInstanceWithContext(ctx, input)
					if err != nil {
						t.Fatalf("error deleting Database Instance: %s", err)
					}
				},
				Config:  testAccInstanceConfig_mariaDB(rName),
				Destroy: true,
			},
		},
	})
}

func TestAccRDSInstance_Storage_maxAllocated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_maxAllocated(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", acctest.Ct10),
				),
			},
			{
				Config: testAccInstanceConfig_Storage_maxAllocated(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", acctest.Ct0),
				),
			},
			{
				Config: testAccInstanceConfig_Storage_maxAllocated(rName, 15),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "15"),
				),
			},
			{
				Config: testAccInstanceConfig_Storage_maxAllocated(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRDSInstance_password(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			// Password should not be shown in error message
			{
				Config:      testAccInstanceConfig_password(rName, "invalid"),
				ExpectError: regexache.MustCompile(`MasterUserPassword is not a valid password because it is shorter than 8 characters`),
			},
			{
				Config: testAccInstanceConfig_password(rName, "valid-password-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "valid-password-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_password(rName, "valid-password-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "valid-password-2"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ManageMasterPassword_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_manageMasterPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_status"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"manage_master_user_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_ManageMasterPassword_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_manageMasterPasswordKMSKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_status"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"manage_master_user_password",
					"master_user_secret_kms_key_id",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_ManageMasterPassword_convertToManaged(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_password(rName, "valid-password"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbCluster1),
					resource.TestCheckNoResourceAttr(resourceName, "manage_master_user_password"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_manageMasterPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbCluster2),
					resource.TestCheckResourceAttrSet(resourceName, "manage_master_user_password"),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "replicate_source_db", sourceResourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttrPair(resourceName, "db_name", sourceResourceName, "db_name"),
					resource.TestCheckResourceAttr(resourceName, "dedicated_log_volume", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUsername, sourceResourceName, names.AttrUsername),

					resource.TestCheckResourceAttr(sourceResourceName, "replicas.#", acctest.Ct0), // Before refreshing source, it will not be aware of replicas
				),
			},
			{
				// Confirm that `replicas` is populated after refreshing source
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(sourceResourceName, "replicas.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_promote(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "replicate_source_db", ""),
					resource.TestCheckResourceAttrPair(resourceName, "db_name", sourceResourceName, "db_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUsername, sourceResourceName, names.AttrUsername),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const identifierPrefix = "tf-acc-test-prefix-"
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_namePrefix(identifierPrefix, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrIdentifier, identifierPrefix),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", identifierPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_nameGenerated(sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_addLater(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_addLaterSetup(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
				),
			},
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_addLater(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_allocatedStorage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_allocatedStorage(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, acctest.Ct10),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_iops(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_iops(rName, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "1000"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_allocatedStorageAndIops(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_allocatedStorageAndIOPS(rName, 220, 2200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "220"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "2200"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_allowMajorVersionUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_allowMajorVersionUpgrade(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_autoMinorVersionUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_autoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_availabilityZone(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_availabilityZone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_backupRetentionPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_backupRetentionPeriod(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_backupWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_backupWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_window", "00:00-08:00"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_dbSubnetGroupName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_dbSubnetGroupName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_dbSubnetGroupNameRAMShared(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternateAccountAndAlternateRegion(ctx, t),
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_DBSubnetGroupName_ramShared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_dbSubnetGroupNameVPCSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_DBSubnetGroupName_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_deletionProtection(t *testing.T) {
	acctest.Skip(t, "CreateDBInstanceReadReplica API currently ignores DeletionProtection=true with SourceDBInstanceIdentifier set")
	// --- FAIL: TestAccRDSInstance_ReplicateSourceDB_deletionProtection (1624.88s)
	//     testing.go:527: Step 0 error: Check failed: Check 4/4 error: aws_db_instance.test: Attribute 'deletion_protection' expected "true", got "false"
	//
	// Action=CreateDBInstanceReadReplica&AutoMinorVersionUpgrade=true&CopyTagsToSnapshot=false&DBInstanceClass=db.t2.micro&DBInstanceIdentifier=tf-acc-test-6591588621809891413&DeletionProtection=true&PubliclyAccessible=false&SourceDBInstanceIdentifier=tf-acc-test-6591588621809891413-source&Tags=&Version=2014-10-31
	// <RestoreDBInstanceFromDBSnapshotResponse xmlns="http://rds.amazonaws.com/doc/2014-10-31/">
	//   <RestoreDBInstanceFromDBSnapshotResult>
	//     <DBInstance>
	//       <DeletionProtection>false</DeletionProtection>
	//
	// AWS Support has confirmed this issue and noted that it will be fixed in the future.

	ctx := acctest.Context(t)

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_deletionProtection(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_deletionProtection(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_iamDatabaseAuthenticationEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_iamDatabaseAuthenticationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_maintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_maintenanceWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "sun:23:00-sun:23:30"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_maxAllocatedStorage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_maxAllocatedStorage(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", acctest.Ct10),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_monitoring(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_monitoring(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "5"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_monitoring_sourceAlreadyExists(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_monitoring_sourceOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
				),
			},
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_monitoring(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "5"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_multiAZ(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_multiAZ(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_networkType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_networkType(rName, "IPV4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_parameterGroupNameSameSetOnBoth(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_sameSetOnBoth(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, sourceResourceName, names.AttrParameterGroupName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_parameterGroupNameDifferentSetOnBoth(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_differentSetOnBoth(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					resource.TestCheckResourceAttr(sourceResourceName, names.AttrParameterGroupName, fmt.Sprintf("%s-source", rName)),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, rName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_parameterGroupNameReplicaCopiesValue(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_replicaCopiesValue(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, sourceResourceName, names.AttrParameterGroupName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_parameterGroupNameSetOnReplica(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_setOnReplica(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, rName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_port(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_port(rName, 9999),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "9999"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_vpcSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_caCertificateIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"
	certifiateDataSourceName := "data.aws_rds_certificate.latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_caCertificateID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttrPair(sourceResourceName, "ca_cert_identifier", certifiateDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", certifiateDataSourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_characterSet_Source(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_characterSet_Source(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					resource.TestCheckResourceAttr(sourceResourceName, "character_set_name", "WE8ISO8859P15"),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "replicate_source_db", sourceResourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, "character_set_name", "WE8ISO8859P15"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"manage_master_user_password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_characterSet_Replica(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_characterSet_Replica(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					resource.TestCheckResourceAttr(sourceResourceName, "character_set_name", "WE8ISO8859P15"),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "replicate_source_db", sourceResourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, "character_set_name", "WE8ISO8859P15"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_replicaMode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_replicaMode(rName, rds.ReplicaModeMounted),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "replica_mode", rds.ReplicaModeMounted),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"manage_master_user_password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_replicaMode(rName, rds.ReplicaModeOpenReadOnly),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "replica_mode", rds.ReplicaModeOpenReadOnly),
				),
			},
		},
	})
}

// When an RDS Instance is added in a separate apply from the creation of the
// source instance, and the parameter group is changed on the replica, it can
// sometimes lead to the API trying to reboot the instance when another
// "management operation" is in progress:
//
// InvalidDBInstanceState: Instance cannot currently reboot due to an in-progress management operation
// https://github.com/hashicorp/terraform-provider-aws/issues/11905
func TestAccRDSInstance_ReplicateSourceDB_parameterGroupTwoStep(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	sourceResourceName := "aws_db_instance.source"
	parameterGroupResourceName := "aws_db_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_ParameterGroupTwoStep_setup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					resource.TestCheckResourceAttr(sourceResourceName, names.AttrParameterGroupName, "default.oracle-ee-19"),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_parameterGroupTwoStep(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					resource.TestCheckResourceAttr(sourceResourceName, names.AttrParameterGroupName, "default.oracle-ee-19"),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "replica_mode", "open-read-only"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, parameterGroupResourceName, names.AttrName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_CrossRegion_parameterGroupNameEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_CrossRegion_ParameterGroupName_equivalent(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExistsWithProvider(ctx, sourceResourceName, &sourceDbInstance, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
					resource.TestCheckResourceAttr(sourceResourceName, names.AttrParameterGroupName, fmt.Sprintf("%s-source", rName)),
					testAccCheckDBInstanceExistsWithProvider(ctx, resourceName, &dbInstance, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					resource.TestCheckResourceAttrPair(resourceName, "replicate_source_db", sourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, "aws_db_parameter_group.test", names.AttrName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_CrossRegion_parameterGroupNamePostgres(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_CrossRegion_ParameterGroupName_postgres(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExistsWithProvider(ctx, sourceResourceName, &sourceDbInstance, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
					resource.TestCheckResourceAttr(sourceResourceName, names.AttrParameterGroupName, fmt.Sprintf("%s-source", rName)),
					testAccCheckDBInstanceExistsWithProvider(ctx, resourceName, &dbInstance, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					resource.TestCheckResourceAttrPair(resourceName, "replicate_source_db", sourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, "aws_db_parameter_group.test", names.AttrName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
					testAccCheckInstanceParameterApplyStatusInSync(&sourceDbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_CrossRegion_characterSet(t *testing.T) {
	t.Skip("Skipping due to upstream error")
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_CrossRegion_CharacterSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExistsWithProvider(ctx, sourceResourceName, &sourceDbInstance, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
					resource.TestCheckResourceAttr(sourceResourceName, "character_set_name", "WE8ISO8859P15"),
					testAccCheckDBInstanceExistsWithProvider(ctx, resourceName, &dbInstance, acctest.RegionProviderFunc(acctest.Region(), &providers)),
					resource.TestCheckResourceAttrPair(resourceName, "replicate_source_db", sourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "character_set_name", "WE8ISO8859P15"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_s3Import(t *testing.T) {
	acctest.Skip(t, "RestoreDBInstanceFromS3 cannot restore from MySQL version 5.6")

	ctx := acctest.Context(t)

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_s3Import(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_snapshotID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "dedicated_log_volume", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", sourceDbResourceName, "instance_class"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAllocatedStorage, sourceDbResourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, sourceDbResourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, sourceDbResourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUsername, sourceDbResourceName, names.AttrUsername),
					resource.TestCheckResourceAttrPair(resourceName, "db_name", sourceDbResourceName, "db_name"),
					resource.TestCheckResourceAttrPair(resourceName, "maintenance_window", sourceDbResourceName, "maintenance_window"),
					resource.TestCheckResourceAttrPair(resourceName, "option_group_name", sourceDbResourceName, "option_group_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, sourceDbResourceName, names.AttrParameterGroupName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPort, sourceDbResourceName, names.AttrPort),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_ManageMasterPasswordKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_snapshotID_ManageMasterPasswordKMSKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_status"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					"manage_master_user_password",
					"master_user_secret_kms_key_id",
					"snapshot_identifier",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const identifierPrefix = "tf-acc-test-prefix-"
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotIdentifier_namePrefix(identifierPrefix, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrIdentifier, identifierPrefix),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", identifierPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	sourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const resourceName = "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotIdentifier_nameGenerated(sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_AssociationRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance1, dbInstance2 rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_snapshotID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance1),
				),
			},
			{
				Config: testAccInstanceConfig_SnapshotID_associationRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance2),
					testAccCheckDBInstanceNotRecreated(&dbInstance1, &dbInstance2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAllocatedStorage, sourceDbResourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, sourceDbResourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUsername, sourceDbResourceName, names.AttrUsername),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_allocatedStorage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_allocatedStorage(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, acctest.Ct10),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_io1Storage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_ioStorage(rName, "io1", 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "1000"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_io2Storage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_ioStorage(rName, "io2", 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "1000"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_allowMajorVersionUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_allowMajorVersionUpgrade(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllowMajorVersionUpgrade, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_autoMinorVersionUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_autoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_availabilityZone(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_availabilityZone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_backupRetentionPeriodOverride(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_backupRetentionPeriod(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_backupRetentionPeriodUnset(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_BackupRetentionPeriod_unset(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_backupWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_backupWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_window", "00:00-08:00"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_dbSubnetGroupName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_dbSubnetGroupName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_dbSubnetGroupNameRAMShared(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_DBSubnetGroupName_ramShared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_dbSubnetGroupNameVPCSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_DBSubnetGroupName_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_deletionProtection(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccInstanceConfig_SnapshotID_deletionProtection(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_iamDatabaseAuthenticationEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_iamDatabaseAuthenticationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_maintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_maintenanceWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "sun:23:00-sun:23:30"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_maxAllocatedStorage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_maxAllocatedStorage(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", acctest.Ct10),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_monitoring(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_monitoring(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "5"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_multiAZ(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_multiAZ(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_multiAZSQLServer(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_MultiAZ_sqlServer(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "listener_endpoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multi_az", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_parameterGroupName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_parameterGroupName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, rName),
					testAccCheckInstanceParameterApplyStatusInSync(&dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_port(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_port(rName, 9999),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "9999"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_tagsRemove(t *testing.T) {
	acctest.Skip(t, "To be fixed: https://github.com/hashicorp/terraform-provider-aws/issues/26808")
	// --- FAIL: TestAccRDSInstance_SnapshotIdentifierTags_unset (1086.15s)
	//     testing.go:527: Step 0 error: Check failed: Check 4/4 error: aws_db_instance.test: Attribute 'tags.%' expected "0", got "1"

	ctx := acctest.Context(t)

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_tagsRemove(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_vpcSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
		},
	})
}

// Regression reference: https://github.com/hashicorp/terraform-provider-aws/issues/5360
// This acceptance test explicitly tests when snapshot_identifier is set,
// vpc_security_group_ids is set (which triggered the resource update function),
// and tags is set which was missing its ARN used for tagging
func TestAccRDSInstance_SnapshotIdentifier_vpcSecurityGroupIDsTags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_VPCSecurityGroupIDs_tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccRDSInstance_monitoringInterval(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "60"),
				),
			},
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", acctest.Ct0),
				),
			},
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MonitoringRoleARN_enabledToDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_monitoringInterval(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRDSInstance_MonitoringRoleARN_enabledToRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_monitoringRoleARNRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccRDSInstance_MonitoringRoleARN_removedToEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_monitoringRoleARNRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, names.AttrARN),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/3760 .
// We apply a plan, then change just the iops. If the apply succeeds, we
// consider this a pass, as before in 3760 the request would fail
func TestAccRDSInstance_Storage_separateIOPSUpdate_Io1(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_iopsUpdate(rName, "io1", 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
				),
			},

			{
				Config: testAccInstanceConfig_iopsUpdate(rName, "io1", 2000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
				),
			},
		},
	})
}

func TestAccRDSInstance_Storage_separateIOPSUpdate_Io2(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_iopsUpdate(rName, "io2", 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
				),
			},

			{
				Config: testAccInstanceConfig_iopsUpdate(rName, "io2", 2000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
				),
			},
		},
	})
}

func TestAccRDSInstance_portUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mySQLPort(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "3306"),
				),
			},

			{
				Config: testAccInstanceConfig_updateMySQLPort(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "3305"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MSSQL_tz(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_MSSQL_timezone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes_MSSQL(&v, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "20"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.InstanceEngineSQLServerExpress),
				),
			},

			{
				Config: testAccInstanceConfig_MSSQL_timezone_AKST(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes_MSSQL(&v, "Alaskan Standard Time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "20"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfrds.InstanceEngineSQLServerExpress),
				),
			},
		},
	})
}

func TestAccRDSInstance_MSSQL_domain(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vBefore, vAfter rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := acctest.RandomDomain()
	domain1 := domain.RandomSubdomain().String()
	domain2 := domain.RandomSubdomain().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mssqlDomain(rName, domain1, domain2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &vBefore),
					testAccCheckInstanceDomainAttributes(domain1, &vBefore),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomain),
					resource.TestCheckResourceAttrSet(resourceName, "domain_iam_role_name"),
				),
			},
			{
				Config: testAccInstanceConfig_mssqlUpdateDomain(rName, domain1, domain2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &vAfter),
					testAccCheckInstanceDomainAttributes(domain2, &vAfter),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomain),
					resource.TestCheckResourceAttrSet(resourceName, "domain_iam_role_name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MSSQL_domainSnapshotRestore(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v, vRestoredInstance rds.DBInstance
	resourceName := "aws_db_instance.test"
	originResourceName := "aws_db_instance.origin"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mssqlDomainSnapshotRestore(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &vRestoredInstance),
					testAccCheckDBInstanceExists(ctx, originResourceName, &v),
					testAccCheckInstanceDomainAttributes(domain, &vRestoredInstance),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomain),
					resource.TestCheckResourceAttrSet(resourceName, "domain_iam_role_name"),
				),
			},
		},
	})
}

func TestAccRDSInstance_MSSQL_selfManagedDomain(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vBefore, vAfter rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomain().String()
	domainOu := fmt.Sprintf("OU=AWS,DC=%s,DC=%s", strings.Split(domain, ".")[0], strings.Split(domain, ".")[1])
	domain1 := acctest.RandomDomain().String()
	domain1Ou := fmt.Sprintf("OU=AWS,DC=%s,DC=%s", strings.Split(domain1, ".")[0], strings.Split(domain1, ".")[1])

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mssqlSelfManagedDomain(rName, domain, domainOu),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &vBefore),
					resource.TestCheckResourceAttrSet(resourceName, "domain_fqdn"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_ou"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_auth_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_dns_ips.#", acctest.Ct2),
				),
			},
			{
				Config: testAccInstanceConfig_mssqlUpdateSelfManagedDomain(rName, domain1, domain1Ou),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &vAfter),
					resource.TestCheckResourceAttrSet(resourceName, "domain_fqdn"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_ou"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_auth_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_dns_ips.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccRDSInstance_MSSQL_selfManagedDomainSnapshotRestore(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v, vRestoredInstance rds.DBInstance
	resourceName := "aws_db_instance.test"
	originResourceName := "aws_db_instance.origin"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	domainOu := fmt.Sprintf("OU=AWS,DC=%s,DC=%s", strings.Split(domain, ".")[0], strings.Split(domain, ".")[1])

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mssqlSelfManagedDomainSnapshotRestore(rName, domain, domainOu),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &vRestoredInstance),
					testAccCheckDBInstanceExists(ctx, originResourceName, &v),
					testAccCheckInstanceDomainAttributes(domain, &vRestoredInstance),
					resource.TestCheckResourceAttrSet(resourceName, "domain_fqdn"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_ou"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_auth_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_dns_ips.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccRDSInstance_MySQL_snapshotRestoreWithEngineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v, vRestoredInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	restoreResourceName := "aws_db_instance.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_mySQLSnapshotRestoreEngineVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, restoreResourceName, &vRestoredInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					// Hardcoded older version. Will need to update when no longer compatible to upgrade from this to the default version.
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "8.0.31"),
					resource.TestCheckResourceAttrPair(restoreResourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.default", names.AttrVersion),
				),
			},
		},
	})
}

func TestAccRDSInstance_Versions_minor(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_autoMinorVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, "aws_db_instance.bar", &v),
				),
			},
		},
	})
}

func TestAccRDSInstance_CloudWatchLogsExport_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cloudWatchLogsExport(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_CloudWatchLogsExport_db2(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	// Requires an IBM Db2 License set as environmental variable.
	// Licensing pre-requisite: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/db2-licensing.html.
	customerID := acctest.SkipIfEnvVarNotSet(t, "RDS_DB2_CUSTOMER_ID")
	siteID := acctest.SkipIfEnvVarNotSet(t, "RDS_DB2_SITE_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_CloudWatchLogsExport_db2(rName, customerID, siteID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_CloudWatchLogsExport_mySQL(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cloudWatchLogsExport(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
				),
			},
			{
				Config: testAccInstanceConfig_cloudWatchLogsExportAdd(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "general"),
				),
			},
			{
				Config: testAccInstanceConfig_cloudWatchLogsExportModify(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "general"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "slowquery"),
				),
			},
			{
				Config: testAccInstanceConfig_cloudWatchLogsExportDelete(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRDSInstance_CloudWatchLogsExport_msSQL(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_CloudWatchLogsExport_mssql(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_CloudWatchLogsExport_oracle(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_CloudWatchLogsExport_oracle(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_CloudWatchLogsExport_postgresql(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_CloudWatchLogsExport_postgreSQL(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_dedicatedLogVolume_enableOnCreate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_dedicatedLogVolumeEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "dedicated_log_volume", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_dedicatedLogVolume_enableOnUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_dedicatedLogVolumeEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "dedicated_log_volume", acctest.CtFalse),
				),
			},
			{
				Config: testAccInstanceConfig_dedicatedLogVolumeEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "dedicated_log_volume", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
			{
				Config: testAccInstanceConfig_dedicatedLogVolumeEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "dedicated_log_volume", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccRDSInstance_noDeleteAutomatedBackups(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceAutomatedBackupsDelete(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_noDeleteAutomatedBackups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8792
func TestAccRDSInstance_PerformanceInsights_disabledToEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, tfrds.InstanceEngineMySQL)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_performanceInsightsDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
					"skip_final_snapshot",
					names.AttrFinalSnapshotIdentifier,
				},
			},
			{
				Config: testAccInstanceConfig_performanceInsightsEnabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_PerformanceInsights_enabledToDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, tfrds.InstanceEngineMySQL)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_performanceInsightsEnabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
					"skip_final_snapshot",
					names.AttrFinalSnapshotIdentifier,
				},
			},
			{
				Config: testAccInstanceConfig_performanceInsightsDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSInstance_PerformanceInsights_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, tfrds.InstanceEngineMySQL)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_performanceInsightsKMSKeyID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
					"skip_final_snapshot",
					names.AttrFinalSnapshotIdentifier,
				},
			},
			{
				Config: testAccInstanceConfig_performanceInsightsKMSKeyIdDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccInstanceConfig_performanceInsightsKMSKeyID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccRDSInstance_PerformanceInsights_retentionPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, tfrds.InstanceEngineMySQL)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_performanceInsightsRetentionPeriod(rName, 731),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "731"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
					"skip_final_snapshot",
					names.AttrFinalSnapshotIdentifier,
				},
			},
			{
				Config: testAccInstanceConfig_performanceInsightsRetentionPeriod(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
			{
				Config: testAccInstanceConfig_performanceInsightsRetentionPeriod(rName, 155),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "155"),
				),
			},
		},
	})
}

func TestAccRDSInstance_ReplicateSourceDB_performanceInsightsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, tfrds.InstanceEngineMySQL)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ReplicateSourceDB_performanceInsightsEnabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceResourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					testAccCheckInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccRDSInstance_SnapshotIdentifier_performanceInsightsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, tfrds.InstanceEngineMySQL)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_SnapshotID_performanceInsightsEnabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccRDSInstance_caCertificateIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	dataSourceName := "data.aws_rds_certificate.latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_caCertificateID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", dataSourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccRDSInstance_RestoreToPointInTime_sourceIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_RestoreToPointInTime_sourceID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"delete_automated_backups",
					names.AttrFinalSnapshotIdentifier,
					"latest_restorable_time", // dynamic value of a DBInstance
					names.AttrPassword,
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_RestoreToPointInTime_sourceResourceID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_RestoreToPointInTime_sourceResourceID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"delete_automated_backups",
					names.AttrFinalSnapshotIdentifier,
					"latest_restorable_time", // dynamic value of a DBInstance
					names.AttrPassword,
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_RestoreToPointInTime_monitoring(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_RestoreToPointInTime_monitoring(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "5"),
				),
			},
		},
	})
}

func TestAccRDSInstance_RestoreToPointInTime_manageMasterPassword(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance, sourceDbInstance rds.DBInstance
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_RestoreToPointInTime_ManageMasterPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "manage_master_user_password", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "master_user_secret.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "master_user_secret.0.secret_status"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"delete_automated_backups",
					names.AttrFinalSnapshotIdentifier,
					"latest_restorable_time", // dynamic value of a DBInstance
					"manage_master_user_password",
					names.AttrPassword,
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_Oracle_nationalCharacterSet(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Oracle_nationalCharacterSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "nchar_character_set_name", "UTF8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_Oracle_noNationalCharacterSet(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbInstance rds.DBInstance

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Oracle_noNationalCharacterSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "nchar_character_set_name", "AL16UTF16"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccRDSInstance_Outposts_coIPEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, true, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttr(
						resourceName, "customer_owned_ip_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_Outposts_coIPDisabledToEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, false, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
					"skip_final_snapshot",
					names.AttrFinalSnapshotIdentifier,
				},
			},
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, true, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_Outposts_coIPEnabledToDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, true, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrPassword,
					"skip_final_snapshot",
					names.AttrFinalSnapshotIdentifier,
				},
			},
			{
				Config: testAccInstanceConfig_Outpost_coIPEnabled(rName, false, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSInstance_Outposts_coIPRestoreToPointInTime(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance, sourceDbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outposts_coIPRestorePointInTime(rName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceName, &sourceDbInstance),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"delete_automated_backups",
					names.AttrFinalSnapshotIdentifier,
					"latest_restorable_time", // dynamic value of a DBInstance
					names.AttrPassword,
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_Outposts_coIPSnapshotIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot types.DBSnapshot

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDbResourceName := "aws_db_instance.test"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outposts_coIPSnapshotID(rName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, sourceDbResourceName, &sourceDbInstance),
					testAccCheckDBSnapshotExists(ctx, snapshotResourceName, &dbSnapshot),
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ip_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_Outposts_backupTarget(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Outposts_backupTarget(rName, "outposts", 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					testAccCheckInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "backup_target", "outposts"),
				),
			},
		},
	})
}

func TestAccRDSInstance_license(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_license(rName, "license-included"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "license_model", "license-included"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
			{
				Config: testAccInstanceConfig_license(rName, "bring-your-own-license"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "license_model", "bring-your-own-license"),
				),
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_updateEngineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_engineVersion(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.initial", names.AttrVersion),
				),
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_engineVersion(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.update", names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, "blue_green_update.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_updateParameterGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	parameterGroupResourceName := "aws_db_parameter_group.test"
	parameterGroupDataSource := "data.aws_db_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, parameterGroupDataSource, names.AttrName),
				),
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_parameterGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, parameterGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "blue_green_update.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"latest_restorable_time", // This causes intermittent failures when the value increments
					"blue_green_update",
				},
			},
		},
	})
}

// Updating tags should bypass the Blue/Green Deployment
func TestAccRDSInstance_BlueGreenDeployment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_tags1(rName, acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_updateInstanceClass(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_updateableInstanceClass(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
				),
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_updateableInstanceClass(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_update.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_updateAndPromoteReplica(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	sourceResourceName := "aws_db_instance.source"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_prePromote(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "replicate_source_db", sourceResourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
				),
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_promote(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.update", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_update.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"blue_green_update",
					"delete_automated_backups",
					names.AttrFinalSnapshotIdentifier,
					"latest_restorable_time",
					names.AttrPassword,
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_updateAndEnableBackups(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_pre(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct0),
				),
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_updateableInstanceClass(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_update.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_deletionProtectionBypassesBlueGreen(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_deletionProtection(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_deletionProtection(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "blue_green_update.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_passwordBypassesBlueGreen(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_password(rName, "valid-password-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "valid-password-1"),
				),
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_password(rName, "valid-password-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "valid-password-2"),
				),
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_updateWithDeletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_deletionProtection(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_deletionProtection(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_update.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
			{
				Config: testAccInstanceConfig_BlueGreenDeployment_deletionProtection(rName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v3),
					testAccCheckDBInstanceNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "blue_green_update.0.enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSInstance_BlueGreenDeployment_outOfBand(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"
	var updateVersion string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_engineVersion(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, "data.aws_rds_engine_version.initial", names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrResourceID),
					testAccCheckRetrieveValue("data.aws_rds_engine_version.update", names.AttrVersion, &updateVersion),
				),
			},
			{
				PreConfig: func() {
					meta := acctest.Provider.Meta().(*conns.AWSClient)
					conn := meta.RDSClient(ctx)
					deadline := tfresource.NewDeadline(40 * time.Minute)

					orchestrator := tfrds.NewBlueGreenOrchestrator(conn)
					defer orchestrator.CleanUp(ctx)

					input := &rds_sdkv2.CreateBlueGreenDeploymentInput{
						BlueGreenDeploymentName: aws.String(rName),
						Source:                  v1.DBInstanceArn,
						TargetEngineVersion:     aws.String(updateVersion),
					}

					dep, err := orchestrator.CreateDeployment(ctx, input)
					if err != nil {
						t.Fatalf("creating Blue/Green Deployment: %s", err)
					}

					deploymentIdentifier := dep.BlueGreenDeploymentIdentifier

					defer func() {
						// Ensure that the Blue/Green Deployment is always cleaned up
						input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
							BlueGreenDeploymentIdentifier: deploymentIdentifier,
						}
						if aws.StringValue(dep.Status) != "SWITCHOVER_COMPLETED" {
							input.DeleteTarget = aws.Bool(true)
						}
						_, err = conn.DeleteBlueGreenDeployment(ctx, input)
						if err != nil {
							t.Fatalf("deleting Blue/Green Deployment: %s", err)
						}

						orchestrator.AddCleanupWaiter(func(ctx context.Context, conn *rds_sdkv2.Client, optFns ...tfresource.OptionsFunc) {
							_, err = tfrds.WaitBlueGreenDeploymentDeleted(ctx, conn, aws.StringValue(deploymentIdentifier), deadline.Remaining(), optFns...)
							if err != nil {
								t.Fatalf("waiting for Blue/Green Deployment to be deleted: %s", err)
							}
						})
					}()

					dep, err = tfrds.WaitBlueGreenDeploymentAvailable(ctx, conn, aws.StringValue(deploymentIdentifier), deadline.Remaining())
					if err != nil {
						t.Fatalf("waiting for Blue/Green Deployment to be available: %s", err)
					}
					targetARN, err := tfrds.ParseDBInstanceARN(aws.StringValue(dep.Target))
					if err != nil {
						t.Fatalf("parsing target ARN: %s", err)
					}
					_, err = tfrds.WaitDBInstanceAvailable(ctx, conn, targetARN.Identifier, deadline.Remaining())
					if err != nil {
						t.Fatalf("waiting for Green instance to be available: %s", err)
					}

					dep, err = orchestrator.Switchover(ctx, aws.StringValue(dep.BlueGreenDeploymentIdentifier), deadline.Remaining())
					if err != nil {
						t.Fatalf("switching over: %s", err)
					}

					sourceARN, err := tfrds.ParseDBInstanceARN(aws.StringValue(dep.Source))
					if err != nil {
						t.Fatalf("parsing source ARN: %s", err)
					}

					deleteInput := &rds_sdkv2.DeleteDBInstanceInput{
						DBInstanceIdentifier: aws.String(sourceARN.Identifier),
						SkipFinalSnapshot:    aws.Bool(true),
					}
					_, err = tfresource.RetryWhen(ctx, 5*time.Minute,
						func() (any, error) {
							return conn.DeleteDBInstance(ctx, deleteInput)
						},
						func(err error) (bool, error) {
							// Retry for IAM eventual consistency.
							if tfawserr_sdkv2.ErrMessageContains(err, tfrds.ErrCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
								return true, err
							}

							if tfawserr_sdkv2.ErrMessageContains(err, tfrds.ErrCodeInvalidParameterCombination, "disable deletion pro") {
								return true, err
							}

							return false, err
						},
					)
					if err != nil {
						t.Fatalf("deleting source instance: %s", err)
					}

					orchestrator.AddCleanupWaiter(func(ctx context.Context, conn *rds_sdkv2.Client, optFns ...tfresource.OptionsFunc) {
						_, err = tfrds.WaitDBInstanceDeleted(ctx, meta.RDSConn(ctx), sourceARN.Identifier, deadline.Remaining(), optFns...)
						if err != nil {
							t.Fatalf("waiting for source instance to be deleted: %s", err)
						}
					})
				},
				Config: testAccInstanceConfig_engineVersion(rName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrResourceID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
					"latest_restorable_time",
				},
			},
		},
	})
}

func TestAccRDSInstance_Storage_gp3MySQL(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_gp3(rName, testAccInstanceConfig_orderableClassMySQLGP3, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
				},
			},
			{
				Config: testAccInstanceConfig_Storage_gp3(rName, testAccInstanceConfig_orderableClassMySQLGP3, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "300"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
		},
	})
}

func TestAccRDSInstance_Storage_gp3Postgres(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_gp3(rName, testAccInstanceConfig_orderableClassPostgresGP3, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
				},
			},
			{
				Config: testAccInstanceConfig_Storage_gp3(rName, testAccInstanceConfig_orderableClassPostgresGP3, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "300"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
		},
	})
}

func TestAccRDSInstance_Storage_gp3SQLServer(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_gp3(rName, testAccInstanceConfig_orderableClassSQLServerExGP3, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
				},
			},
			{
				Config: testAccInstanceConfig_Storage_gp3(rName, testAccInstanceConfig_orderableClassSQLServerExGP3, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "300"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
		},
	})
}

// // https://github.com/hashicorp/terraform-provider-aws/issues/33512
func TestAccRDSInstance_Storage_changeThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_throughput(rName, 12000, 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "12000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
			{
				Config: testAccInstanceConfig_Storage_throughput(rName, 12000, 600),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "12000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "600"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/33512
func TestAccRDSInstance_Storage_changeIOPSThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_throughput(rName, 12000, 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "12000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
			{
				Config: testAccInstanceConfig_Storage_throughput(rName, 13000, 600),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "13000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "600"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/33512
func TestAccRDSInstance_Storage_changeIOPS(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_throughput(rName, 12000, 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "12000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
			{
				Config: testAccInstanceConfig_Storage_throughput(rName, 13000, 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "13000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/33512
func TestAccRDSInstance_Storage_throughputSSE(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_throughputSSE(rName, 4201, 125),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "4201"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
			{
				Config: testAccInstanceConfig_Storage_throughputSSE(rName, 4201, 126),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "4201"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "126"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
		},
	})
}

func TestAccRDSInstance_Storage_typePostgres(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_Storage_typePostgres(rName, "gp2", 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrFinalSnapshotIdentifier,
					names.AttrPassword,
					"skip_final_snapshot",
					"delete_automated_backups",
					"blue_green_update",
				},
			},
			{
				Config: testAccInstanceConfig_Storage_typePostgres(rName, "gp3", 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, "300"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, "storage_throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "gp3"),
				),
			},
		},
	})
}

func TestAccRDSInstance_newIdentifier_Pending(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
				),
			},
			{
				Config:             testAccInstanceConfig_basic(rName2),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						// Change Identifier is a pending change
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
				),
			},
		},
	})
}

func TestAccRDSInstance_newIdentifier_Immediately(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
				),
			},
			{
				Config: testAccInstanceConfig_basicApplyImmediately(rName2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &v2),
					testAccCheckDBInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName2),
				),
			},
		},
	})
}

func testAccCheckInstanceAutomatedBackupsDelete(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_instance" {
				continue
			}

			log.Printf("[INFO] Trying to locate the DBInstance Automated Backup")
			describeOutput, err := conn.DescribeDBInstanceAutomatedBackupsWithContext(ctx, &rds.DescribeDBInstanceAutomatedBackupsInput{
				DBInstanceIdentifier: aws.String(rs.Primary.Attributes[names.AttrIdentifier]),
			})
			if err != nil {
				return err
			}

			if describeOutput == nil || len(describeOutput.DBInstanceAutomatedBackups) == 0 {
				return fmt.Errorf("Automated backup for %s not found", rs.Primary.Attributes[names.AttrIdentifier])
			}

			log.Printf("[INFO] Deleting automated backup for %s", rs.Primary.Attributes[names.AttrIdentifier])
			_, err = conn.DeleteDBInstanceAutomatedBackupWithContext(ctx, &rds.DeleteDBInstanceAutomatedBackupInput{
				DbiResourceId: describeOutput.DBInstanceAutomatedBackups[0].DbiResourceId,
			})
			if err != nil {
				return err
			}
		}

		return testAccCheckDBInstanceDestroy(ctx)(s)
	}
}

func testAccCheckDBInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_instance" {
				continue
			}

			_, err := tfrds.FindDBInstanceByID(ctx, conn, rs.Primary.Attributes[names.AttrIdentifier])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Instance %s still exists", rs.Primary.Attributes[names.AttrIdentifier])
		}

		return nil
	}
}

func testAccCheckRetrieveValue(name, key string, v *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		*v = rs.Primary.Attributes[key]

		return nil
	}
}

func testAccCheckInstanceAttributes(v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.Engine != tfrds.InstanceEngineMySQL {
			return fmt.Errorf("bad engine: %#v", *v.Engine)
		}

		if *v.EngineVersion == "" {
			return fmt.Errorf("bad engine_version: %#v", *v.EngineVersion)
		}

		if *v.BackupRetentionPeriod != 0 {
			return fmt.Errorf("bad backup_retention_period: %#v", *v.BackupRetentionPeriod)
		}

		return nil
	}
}

func testAccCheckInstanceAttributes_MSSQL(v *rds.DBInstance, tz string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.Engine != tfrds.InstanceEngineSQLServerExpress {
			return fmt.Errorf("bad engine: %#v", *v.Engine)
		}

		rtz := ""
		if v.Timezone != nil {
			rtz = *v.Timezone
		}

		if tz != rtz {
			return fmt.Errorf("Expected (%s) Timezone for MSSQL test, got (%s)", tz, rtz)
		}

		return nil
	}
}

func testAccCheckInstanceDomainAttributes(domain string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, dm := range v.DomainMemberships {
			if *dm.FQDN != domain {
				continue
			}

			return nil
		}

		return fmt.Errorf("Domain %s not found in domain memberships", domain)
	}
}

func testAccCheckInstanceParameterApplyStatusInSync(dbInstance *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, dbParameterGroup := range dbInstance.DBParameterGroups {
			parameterApplyStatus := aws.StringValue(dbParameterGroup.ParameterApplyStatus)
			if parameterApplyStatus != "in-sync" {
				id := aws.StringValue(dbInstance.DBInstanceIdentifier)
				parameterGroupName := aws.StringValue(dbParameterGroup.DBParameterGroupName)
				return fmt.Errorf("expected DB Instance (%s) Parameter Group (%s) apply status to be: \"in-sync\", got: %q", id, parameterGroupName, parameterApplyStatus)
			}
		}

		return nil
	}
}

func testAccCheckInstanceReplicaAttributes(source, replica *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if replica.ReadReplicaSourceDBInstanceIdentifier != nil && *replica.ReadReplicaSourceDBInstanceIdentifier != *source.DBInstanceIdentifier {
			return fmt.Errorf("bad source identifier for replica, expected: '%s', got: '%s'", *source.DBInstanceIdentifier, *replica.ReadReplicaSourceDBInstanceIdentifier)
		}

		return nil
	}
}

// testAccCheckInstanceDestroyWithFinalSnapshot verifies that:
// - The DBInstance has been destroyed
// - A DBSnapshot has been produced
// - Tags have been copied to the snapshot
// The snapshot is deleted.
func testAccCheckInstanceDestroyWithFinalSnapshot(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn1 := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)
		conn2 := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_instance" {
				continue
			}

			finalSnapshotID := rs.Primary.Attributes[names.AttrFinalSnapshotIdentifier]
			output, err := tfrds.FindDBSnapshotByID(ctx, conn2, finalSnapshotID)
			if err != nil {
				return err
			}

			tags, err := tfrds.ListTags(ctx, conn1, aws.StringValue(output.DBSnapshotArn))
			if err != nil {
				return err
			}

			if _, ok := tags["Name"]; !ok {
				return fmt.Errorf("Name tag not found")
			}

			_, err = conn2.DeleteDBSnapshot(ctx, &rds_sdkv2.DeleteDBSnapshotInput{
				DBSnapshotIdentifier: aws.String(finalSnapshotID),
			})

			if err != nil {
				return err
			}

			_, err = tfrds.FindDBInstanceByID(ctx, conn1, rs.Primary.Attributes[names.AttrIdentifier])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Instance %s still exists", rs.Primary.Attributes[names.AttrIdentifier])
		}

		return nil
	}
}

// testAccCheckInstanceDestroyWithoutFinalSnapshot verifies that:
// - The DBInstance has been destroyed
// - No DBSnapshot has been produced
func testAccCheckInstanceDestroyWithoutFinalSnapshot(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn1 := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)
		conn2 := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_instance" {
				continue
			}

			finalSnapshotID := rs.Primary.Attributes[names.AttrFinalSnapshotIdentifier]
			_, err := tfrds.FindDBSnapshotByID(ctx, conn2, finalSnapshotID)

			if err != nil {
				if !tfresource.NotFound(err) {
					return err
				}
			} else {
				return fmt.Errorf("RDS DB Snapshot %s exists", finalSnapshotID)
			}

			_, err = tfrds.FindDBInstanceByID(ctx, conn1, rs.Primary.Attributes[names.AttrIdentifier])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Instance %s still exists", rs.Primary.Attributes[names.AttrIdentifier])
		}

		return nil
	}
}

func testAccCheckDBInstanceRecreated(i, j *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if dbInstanceIdentityEqual(i, j) {
			return fmt.Errorf("RDS DB Instance not recreated")
		}
		return nil
	}
}

func testAccCheckDBInstanceNotRecreated(i, j *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !dbInstanceIdentityEqual(i, j) {
			return fmt.Errorf("RDS DB Instance recreated")
		}
		return nil
	}
}

func dbInstanceIdentityEqual(i, j *rds.DBInstance) bool {
	return dbInstanceIdentity(i) == dbInstanceIdentity(j)
}

func dbInstanceIdentity(v *rds.DBInstance) string {
	return aws.StringValue(v.DbiResourceId)
}

func testAccCheckDBInstanceExists(ctx context.Context, n string, v *rds.DBInstance) resource.TestCheckFunc {
	return testAccCheckDBInstanceExistsWithProvider(ctx, n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckDBInstanceExistsWithProvider(ctx context.Context, n string, v *rds.DBInstance, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.Attributes[names.AttrIdentifier] == "" {
			return fmt.Errorf("No RDS DB Instance ID is set")
		}

		conn := providerF().Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindDBInstanceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccInstanceConfig_orderableClass(engine, license, storage string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[4]s]
}
`, engine, license, storage, mainInstanceClasses)
}

func testAccInstanceConfig_orderableClassDB2() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEngineDB2Standard, "bring-your-own-license", "gp3")
}

func testAccInstanceConfig_orderableClassMySQL() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEngineMySQL, "general-public-license", "standard")
}

func testAccInstanceConfig_orderableClassMySQLGP3() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEngineMySQL, "general-public-license", "gp3")
}

func testAccInstanceConfig_orderableClassPostgres() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEnginePostgres, "postgresql-license", "standard")
}

func testAccInstanceConfig_orderableClassPostgresGP3() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEnginePostgres, "postgresql-license", "gp3")
}

func testAccInstanceConfig_orderableClassMariadb() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEngineMariaDB, "general-public-license", "standard")
}

func testAccInstanceConfig_orderableClassSQLServerEx() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEngineSQLServerExpress, "license-included", "standard")
}

func testAccInstanceConfig_orderableClassSQLServerExGP3() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEngineSQLServerExpress, "license-included", "gp3")
}

func testAccInstanceConfig_orderableClassSQLServerSe() string {
	return testAccInstanceConfig_orderableClass(tfrds.InstanceEngineSQLServerStandard, "license-included", "standard")
}

func testAccInstanceConfig_orderableClassCustomSQLServerWeb() string {
	return testAccInstanceConfig_orderableClass("custom-sqlserver-web", "", "gp2")
}

func testAccInstanceConfig_orderableClassOracleEnterprise() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[4]s]
}
`, tfrds.InstanceEngineOracleEnterprise, "bring-your-own-license", "gp2", strings.Replace(mainInstanceClasses, "db.t3.small", "frodo", 1))
}

func testAccInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rName))
}

func testAccInstanceConfig_basicApplyImmediately(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  apply_immediately       = true

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rName))
}

func testAccInstanceConfig_db2engine(rName, customerId, siteId string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassDB2(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = "tf-db2-pg-%[1]s"
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "rds.ibm_customer_id"
    value        = %[2]s
    apply_method = "immediate"
  }
  parameter {
    name         = "rds.ibm_site_id"
    value        = %[3]s
    apply_method = "immediate"
  }
}

resource "aws_db_instance" "test" {
  allocated_storage    = 100
  db_name              = "test"
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  identifier           = %[1]q
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  parameter_group_name = aws_db_parameter_group.test.name
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  skip_final_snapshot  = true
}
`, rName, customerId, siteId))
}

func testAccInstanceConfig_identifierPrefix(identifierPrefix string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier_prefix   = %[1]q
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
}
`, identifierPrefix))
}

func testAccInstanceConfig_identifierGenerated() string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(), `
resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
}
`)
}

func testAccInstanceConfig_engineLifecycleSupport_disabled(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier               = %[1]q
  allocated_storage        = 10
  backup_retention_period  = 0
  engine                   = data.aws_rds_orderable_db_instance.test.engine
  engine_version           = data.aws_rds_orderable_db_instance.test.engine_version
  engine_lifecycle_support = "open-source-rds-extended-support-disabled"
  instance_class           = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                  = "test"
  parameter_group_name     = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot      = true
  password                 = "avoid-plaintext-passwords"
  username                 = "tfacctest"
  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rName))
}

func testAccInstanceConfig_majorVersionOnly(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rName))
}

func testAccInstanceConfig_kmsKeyID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  kms_key_id              = aws_kms_key.test.arn
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  storage_encrypted       = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rName))
}

func testAccInstanceConfig_DBSubnetGroupName_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  db_subnet_group_name = aws_db_subnet_group.test.name
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  backup_retention_period = 0
  apply_immediately       = true
}
`, rName))
}

func testAccInstanceConfig_DBSubnetGroupName_update(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_vpc" "test2" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  count = 2

  vpc_id            = aws_vpc.test2.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test2.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test2" {
  name       = "%[1]s-2"
  subnet_ids = aws_subnet.test2[*].id
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  db_subnet_group_name = aws_db_subnet_group.test2.name
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  backup_retention_period = 0
  apply_immediately       = true

  depends_on = [aws_db_subnet_group.test]
}
`, rName))
}

func testAccInstanceConfig_networkType(rName string, networkType string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "test" {
  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
  network_type            = %[2]q
  apply_immediately       = true
}
`, rName, networkType))
}

func testAccInstanceConfig_optionGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name              = data.aws_rds_orderable_db_instance.test.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
}

resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = aws_db_option_group.test.engine_name
  engine_version      = aws_db_option_group.test.major_engine_version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name             = "test"
  option_group_name   = aws_db_option_group.test.name
  skip_final_snapshot = true
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
}
`, rName))
}

func testAccInstanceConfig_caCertificateID(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  allocated_storage   = 10
  apply_immediately   = true
  ca_cert_identifier  = data.aws_rds_certificate.latest.id
  engine              = data.aws_rds_orderable_db_instance.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name             = "test"
  skip_final_snapshot = true
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
}
`, rName))
}

func testAccInstanceConfig_iamAuth(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = [%[2]s]

  supports_iam_database_authentication = true
}

resource "aws_db_instance" "test" {
  identifier                          = %[3]q
  allocated_storage                   = 10
  engine                              = data.aws_rds_engine_version.default.engine
  engine_version                      = data.aws_rds_engine_version.default.version
  instance_class                      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                             = "test"
  password                            = "avoid-plaintext-passwords"
  username                            = "tfacctest"
  backup_retention_period             = 0
  skip_final_snapshot                 = true
  parameter_group_name                = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  iam_database_authentication_enabled = true
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName)
}

func testAccInstanceConfig_FinalSnapshotID_skipFinalSnapshot(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier = %[1]q

  allocated_storage       = 5
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  backup_retention_period = 1

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"

  skip_final_snapshot       = true
  final_snapshot_identifier = %[1]q
}
`, rName))
}

func testAccInstanceConfig_baseS3Import(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s/mysql-5-6-xtrabackup.tar.gz"
  source = "./testdata/mysql-5-6-xtrabackup.tar.gz"
  etag   = filemd5("./testdata/mysql-5-6-xtrabackup.tar.gz")
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "test" {
  name = %[1]q

  roles = [
    aws_iam_role.test.name,
  ]

  policy_arn = aws_iam_policy.test.arn
}

data "aws_rds_engine_version" "default" {
  engine = %[2]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  # instance class db.t2.micro is not supported for restoring from S3 # TODO: can we search for instances restorable from s3?
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium", "db.t3.medium"]
}
`, rName, tfrds.InstanceEngineMySQL))
}

func testAccInstanceConfig_s3Import(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseS3Import(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier = %[1]q

  allocated_storage          = 5
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  auto_minor_version_upgrade = true
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                    = "test"
  password                   = "avoid-plaintext-passwords"
  username                   = "tfacctest"
  backup_retention_period    = 0

  character_set_name = "WE8ISO8859P15"

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true
  multi_az             = false
  db_subnet_group_name = aws_db_subnet_group.test.id

  s3_import {
    source_engine         = data.aws_rds_orderable_db_instance.test.engine
    source_engine_version = "5.6" # leave at 5.6 until someone makes a new testdata restore file

    bucket_name    = aws_s3_bucket.test.bucket
    bucket_prefix  = %[1]q
    ingestion_role = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccInstanceConfig_finalSnapshotID(rName1, rName2 string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier = %[1]q

  allocated_storage       = 5
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  backup_retention_period = 1

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"

  copy_tags_to_snapshot     = true
  final_snapshot_identifier = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName1, rName2))
}

func testAccInstanceConfig_monitoringInterval(rName string, monitoringInterval int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

data "aws_rds_engine_version" "default" {
  engine = %[2]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = [%[3]s]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = %[4]d
  monitoring_role_arn = aws_iam_role.test.arn
  db_name             = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName, tfrds.InstanceEngineMySQL, mainInstanceClasses, monitoringInterval)
}

func testAccInstanceConfig_monitoringRoleARNRemoved(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = [%[2]s]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %[3]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name             = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName)
}

func testAccInstanceConfig_monitoringRoleARN(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

data "aws_rds_engine_version" "default" {
  engine = %[2]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = [%[3]s]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = 5
  monitoring_role_arn = aws_iam_role.test.arn
  db_name             = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName, tfrds.InstanceEngineMySQL, mainInstanceClasses)
}

func testAccInstanceConfig_baseForPITR(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "baz"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "foo"
}
`, rName))
}

func testAccInstanceConfig_RestoreToPointInTime_sourceID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseForPITR(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = "%[1]s-restore"
  instance_class = aws_db_instance.test.instance_class
  restore_to_point_in_time {
    source_db_instance_identifier = aws_db_instance.test.identifier
    use_latest_restorable_time    = true
  }
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_RestoreToPointInTime_sourceResourceID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseForPITR(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = "%[1]s-restore"
  instance_class = aws_db_instance.test.instance_class
  restore_to_point_in_time {
    source_dbi_resource_id     = aws_db_instance.test.resource_id
    use_latest_restorable_time = true
  }
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_RestoreToPointInTime_monitoring(rName string, monitoringInterval int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseForPITR(rName),
		testAccInstanceConfig_baseMonitoringRole(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier          = "%[1]s-restore"
  instance_class      = aws_db_instance.test.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn

  restore_to_point_in_time {
    source_dbi_resource_id     = aws_db_instance.test.resource_id
    use_latest_restorable_time = true
  }

  skip_final_snapshot = true
}
`, rName, monitoringInterval))
}

func testAccInstanceConfig_RestoreToPointInTime_ManageMasterPassword(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseForPITR(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = "%[1]s-restore"
  instance_class = aws_db_instance.test.instance_class

  restore_to_point_in_time {
    source_db_instance_identifier = aws_db_instance.test.identifier
    use_latest_restorable_time    = true
  }

  skip_final_snapshot         = true
  manage_master_user_password = true
}
`, rName))
}

func testAccInstanceConfig_iopsUpdate(rName string, sType string, iops int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  preferred_instance_classes = [%[2]s]

  storage_type  = %[4]q
  supports_iops = true
}

resource "aws_db_instance" "test" {
  identifier           = %[3]q
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true

  storage_type      = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage = 200
  iops              = %[5]d
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName, sType, iops)
}

func testAccInstanceConfig_mySQLPort(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  port                 = 3306
  allocated_storage    = 10
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rName))
}

func testAccInstanceConfig_updateMySQLPort(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rName))
}

func testAccInstanceConfig_MSSQL_timezone(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerEx(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccInstanceConfig_MSSQL_timezone_AKST(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerEx(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot     = true
  timezone                = "Alaskan Standard Time"
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccInstanceConfig_ServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Action": "sts:AssumeRole",
		"Principal": {
			"Service": "rds.${data.aws_partition.current.dns_suffix}"
		},
		"Effect": "Allow",
		"Sid": ""
		}
	]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attatch-policy" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSDirectoryServiceAccess"
}
`, rName)
}

func testAccInstanceConfig_baseVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name = %[1]q

  subnet_ids = aws_subnet.test[*].id
}
`, rName))
}

func testAccInstanceConfig_baseMSSQLDomain(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerEx(),
		testAccInstanceConfig_baseVPC(rName),
		testAccInstanceConfig_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}

resource "aws_directory_service_directory" "directory" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

data "aws_partition" "current" {}
`, rName, domain))
}

func testAccInstanceConfig_mssqlDomain(rName, domain1, domain2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseMSSQLDomain(rName, domain1),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]

  domain               = aws_directory_service_directory.directory.id
  domain_iam_role_name = aws_iam_role.role.name
}

resource "aws_directory_service_directory" "directory-2" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, rName, domain2))
}

func testAccInstanceConfig_mssqlUpdateDomain(rName, domain1, domain2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseMSSQLDomain(rName, domain1),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  apply_immediately       = true
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]

  domain               = aws_directory_service_directory.directory-2.id
  domain_iam_role_name = aws_iam_role.role.name
}

resource "aws_directory_service_directory" "directory-2" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, rName, domain2))
}

func testAccInstanceConfig_mssqlDomainSnapshotRestore(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseMSSQLDomain(rName, domain),
		fmt.Sprintf(`
resource "aws_db_instance" "origin" {
  allocated_storage   = 20
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
}

resource "aws_db_snapshot" "origin" {
  db_instance_identifier = aws_db_instance.origin.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage       = 20
  apply_immediately       = true
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = aws_db_instance.origin.engine
  engine_version          = aws_db_instance.origin.engine_version
  identifier              = "%[1]s-restore"
  instance_class          = aws_db_instance.origin.instance_class
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]

  domain               = aws_directory_service_directory.directory.id
  domain_iam_role_name = aws_iam_role.role.name

  snapshot_identifier = aws_db_snapshot.origin.id
}
`, rName))
}

func testAccInstanceConfig_baseMSSQLSelfManagedDomain(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerEx(),
		testAccInstanceConfig_baseVPC(rName),
		testAccInstanceConfig_ServiceRole(rName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}

resource "aws_kms_key" "example" {
  description = "Terraform acc test %[1]s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    },
	{
		"Sid": "Allow use of the KMS key on behalf of RDS",
		"Effect": "Allow",
		"Principal": {
			"Service": [
				"rds.amazonaws.com"
			]
		},
		"Action": "kms:Decrypt",
		"Resource": "*"
	}
  ]
}
 POLICY
}

resource "aws_secretsmanager_secret" "example" {
  name       = %[1]q
  kms_key_id = aws_kms_key.example.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement":[{
    "Effect": "Allow",
    "Principal": {
      "Service": "rds.amazonaws.com"
    },
    "Action": "secretsmanager:GetSecretValue",
    "Resource": "*",
    "Condition": {
      "StringEquals": {
        "aws:sourceAccount": "${data.aws_caller_identity.current.account_id}"
      },
      "ArnLike": {
        "aws:sourceArn": "arn:${data.aws_partition.current.partition}:rds:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:db:*"
      }
    }
  }]
}
POLICY
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({ "CUSTOMER_MANAGED_ACTIVE_DIRECTORY_USERNAME" : "Admin", "CUSTOMER_MANAGED_ACTIVE_DIRECTORY_PASSWORD" : "avoid-plaintext-passwords" })
}
`, rName))
}

func testAccInstanceConfig_mssqlSelfManagedDomain(rName, domain, domainOu string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseMSSQLSelfManagedDomain(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]
  domain_fqdn             = %[2]q
  domain_ou               = %[3]q
  domain_auth_secret_arn  = aws_secretsmanager_secret_version.example.arn
  domain_dns_ips          = ["123.124.125.126", "123.124.125.127"]
}
`, rName, domain, domainOu))
}

func testAccInstanceConfig_mssqlUpdateSelfManagedDomain(rName, domain, domainOu string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseMSSQLSelfManagedDomain(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 20
  apply_immediately       = true
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]
  domain_fqdn             = %[2]q
  domain_ou               = %[3]q
  domain_auth_secret_arn  = aws_secretsmanager_secret_version.example.arn
  domain_dns_ips          = ["123.124.125.126", "123.124.125.127"]
}

resource "aws_secretsmanager_secret" "example-2" {
  name       = "%[1]s-2"
  kms_key_id = aws_kms_key.example.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement":[{
    "Effect": "Allow",
    "Principal": {
      "Service": "rds.amazonaws.com"
    },
    "Action": "secretsmanager:GetSecretValue",
    "Resource": "*",
    "Condition": {
      "StringEquals": {
        "aws:sourceAccount": "${data.aws_caller_identity.current.account_id}"
      },
      "ArnLike": {
        "aws:sourceArn": "arn:${data.aws_partition.current.partition}:rds:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:db:*"
      }
    }
  }]
}
POLICY
}

resource "aws_secretsmanager_secret_version" "example-2" {
  secret_id     = aws_secretsmanager_secret.example-2.id
  secret_string = jsonencode({ "CUSTOMER_MANAGED_ACTIVE_DIRECTORY_USERNAME" : "Admin", "CUSTOMER_MANAGED_ACTIVE_DIRECTORY_PASSWORD" : "avoid-plaintext-passwords" })
}
`, rName, domain, domainOu))
}

func testAccInstanceConfig_mssqlSelfManagedDomainSnapshotRestore(rName, domain, domainOu string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseMSSQLSelfManagedDomain(rName),
		fmt.Sprintf(`

resource "aws_db_instance" "origin" {
  allocated_storage   = 20
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
}

resource "aws_db_snapshot" "origin" {
  db_instance_identifier = aws_db_instance.origin.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage       = 20
  apply_immediately       = true
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = aws_db_instance.origin.engine
  engine_version          = aws_db_instance.origin.engine_version
  identifier              = "%[1]s-restore"
  instance_class          = aws_db_instance.origin.instance_class
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]

  domain_fqdn            = %[2]q
  domain_ou              = %[3]q
  domain_auth_secret_arn = aws_secretsmanager_secret_version.example.arn
  domain_dns_ips         = ["123.124.125.126", "123.124.125.127"]

  snapshot_identifier = aws_db_snapshot.origin.id
}
`, rName, domain, domainOu))
}

func testAccInstanceConfig_mySQLSnapshotRestoreEngineVersion(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 20
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = "8.0.31" # test is from older to newer version, update when restore from this to current default version is incompatible
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "restore" {
  allocated_storage       = 20
  apply_immediately       = true
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = "%[1]s-restore"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot     = true
  snapshot_identifier     = aws_db_snapshot.test.id
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  vpc_security_group_ids  = [aws_security_group.test.id]
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccInstanceConfig_Versions_allowMajor(rName string, allowMajorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage           = 10
  allow_major_version_upgrade = %[1]t
  engine                      = data.aws_rds_orderable_db_instance.test.engine
  engine_version              = data.aws_rds_orderable_db_instance.test.engine_version
  identifier                  = %[2]q
  instance_class              = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                     = "baz"
  password                    = "barbarbarbar"
  skip_final_snapshot         = true
  username                    = "foo"
}
`, allowMajorVersionUpgrade, rName))
}

func testAccInstanceConfig_autoMinorVersion(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "bar" {
  identifier          = %[1]q
  allocated_storage   = 10
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  db_name             = "baz"
  password            = "barbarbarbar"
  username            = "foo"
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_cloudWatchLogsExport(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  skip_final_snapshot  = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
  ]
}
`, rName))
}

func testAccInstanceConfig_cloudWatchLogsExportAdd(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  skip_final_snapshot  = true

  apply_immediately = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
    "general",
  ]
}
`, rName))
}

func testAccInstanceConfig_cloudWatchLogsExportModify(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  skip_final_snapshot  = true

  apply_immediately = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "general",
    "slowquery",
  ]
}
`, rName))
}

func testAccInstanceConfig_cloudWatchLogsExportDelete(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rName))
}

func testAccInstanceConfig_mariaDB(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_DBSubnetGroupName_ramShared(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_association" "test" {
  count    = 2
  provider = "awsalternate"

  resource_arn       = aws_subnet.test[count.index].arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_db_instance" "test" {
  allocated_storage      = 5
  db_subnet_group_name   = aws_db_subnet_group.test.name
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  identifier             = %[1]q
  instance_class         = data.aws_rds_orderable_db_instance.test.instance_class
  password               = "avoid-plaintext-passwords"
  username               = "tfacctest"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_DBSubnetGroupName_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "test" {
  allocated_storage      = 5
  db_subnet_group_name   = aws_db_subnet_group.test.name
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  identifier             = %[1]q
  instance_class         = data.aws_rds_orderable_db_instance.test.instance_class
  password               = "avoid-plaintext-passwords"
  username               = "tfacctest"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_deletionProtection(rName string, deletionProtection bool) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  deletion_protection = %[1]t
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[2]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, deletionProtection, rName))
}

func testAccInstanceConfig_CloudWatchLogsExport_db2(rName, customerId, siteId string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassDB2(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = "tf-db2-pg-%[1]s"
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "rds.ibm_customer_id"
    value        = %[2]s
    apply_method = "immediate"
  }
  parameter {
    name         = "rds.ibm_site_id"
    value        = %[3]s
    apply_method = "immediate"
  }
}

resource "aws_db_instance" "test" {
  allocated_storage               = 100
  db_name                         = "test"
  enabled_cloudwatch_logs_exports = ["diag.log", "notify.log"]
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  engine_version                  = data.aws_rds_orderable_db_instance.test.engine_version
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  parameter_group_name            = aws_db_parameter_group.test.name
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, rName, customerId, siteId))
}

func testAccInstanceConfig_CloudWatchLogsExport_oracle(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = %[1]q
  license_model = "bring-your-own-license"
  storage_type  = "standard"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage               = 10
  enabled_cloudwatch_logs_exports = ["alert", "listener", "trace"]
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  identifier                      = %[3]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  license_model                   = "bring-your-own-license"
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, tfrds.InstanceEngineOracleStandard2, mainInstanceClasses, rName)
}

func testAccInstanceConfig_Oracle_nationalCharacterSet(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = %[1]q
  license_model = "bring-your-own-license"
  storage_type  = "standard"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage        = 10
  engine                   = data.aws_rds_orderable_db_instance.test.engine
  identifier               = %[3]q
  instance_class           = data.aws_rds_orderable_db_instance.test.instance_class
  license_model            = "bring-your-own-license"
  nchar_character_set_name = "UTF8"
  password                 = "avoid-plaintext-passwords"
  username                 = "tfacctest"
  skip_final_snapshot      = true
}
`, tfrds.InstanceEngineOracleStandard2, mainInstanceClasses, rName)
}

func testAccInstanceConfig_Oracle_noNationalCharacterSet(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = %[1]q
  license_model = "bring-your-own-license"
  storage_type  = "standard"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[3]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = "bring-your-own-license"
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, tfrds.InstanceEngineOracleStandard2, mainInstanceClasses, rName)
}

func testAccInstanceConfig_CloudWatchLogsExport_mssql(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerSe(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage               = 20
  enabled_cloudwatch_logs_exports = ["agent", "error"]
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  license_model                   = data.aws_rds_orderable_db_instance.test.license_model
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, rName))
}

func testAccInstanceConfig_CloudWatchLogsExport_postgreSQL(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassPostgres(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage               = 10
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  engine                          = data.aws_rds_engine_version.default.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, rName))
}

func testAccInstanceConfig_Storage_maxAllocated(rName string, maxAllocatedStorage int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage     = 5
  engine                = data.aws_rds_orderable_db_instance.test.engine
  identifier            = %[1]q
  instance_class        = data.aws_rds_orderable_db_instance.test.instance_class
  max_allocated_storage = %[2]d
  password              = "avoid-plaintext-passwords"
  username              = "tfacctest"
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccInstanceConfig_password(rName, password string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = %[2]q
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName, password))
}

func testAccInstanceConfig_manageMasterPassword(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage           = 5
  backup_retention_period     = 0
  engine                      = data.aws_rds_orderable_db_instance.test.engine
  engine_version              = data.aws_rds_orderable_db_instance.test.engine_version
  identifier                  = %[1]q
  instance_class              = data.aws_rds_orderable_db_instance.test.instance_class
  manage_master_user_password = true
  skip_final_snapshot         = true
  username                    = "tfacctest"
}
`, rName))
}

func testAccInstanceConfig_manageMasterPasswordKMSKey(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_kms_key" "example" {
  description = "Terraform acc test %[1]s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
 POLICY

}

resource "aws_db_instance" "test" {
  allocated_storage             = 5
  engine                        = data.aws_rds_orderable_db_instance.test.engine
  identifier                    = %[1]q
  instance_class                = data.aws_rds_orderable_db_instance.test.instance_class
  manage_master_user_password   = true
  master_user_secret_kms_key_id = aws_kms_key.example.arn
  username                      = "tfacctest"
  skip_final_snapshot           = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_promote(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_namePrefix(identifierPrefix, sourceName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier_prefix   = %[2]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, sourceName, identifierPrefix))
}

func testAccInstanceConfig_ReplicateSourceDB_nameGenerated(sourceName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, sourceName))
}

func testAccInstanceConfig_ReplicateSourceDB_addLaterSetup() string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(), `
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}
`)
}

func testAccInstanceConfig_ReplicateSourceDB_addLater(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_ReplicateSourceDB_addLaterSetup(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_allocatedStorage(rName string, allocatedStorage int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = %[2]d
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allocated_storage   = %[2]d
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName, allocatedStorage))
}

func testAccInstanceConfig_ReplicateSourceDB_iops(rName string, iops int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  preferred_instance_classes = [%[2]s]

  storage_type  = "io1"
  supports_iops = true
}

resource "aws_db_instance" "source" {
  allocated_storage       = 200
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[3]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
  iops                    = 1100
  storage_type            = "io1"
}

resource "aws_db_instance" "test" {
  identifier          = %[3]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
  iops                = %[4]d
  storage_type        = "io1"
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName, iops)
}

func testAccInstanceConfig_ReplicateSourceDB_allocatedStorageAndIOPS(rName string, allocatedStorage, iops int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  preferred_instance_classes = [%[2]s]

  storage_type  = "io1"
  supports_iops = true
}

resource "aws_db_instance" "source" {
  allocated_storage       = %[3]d
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[4]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
  iops                    = 1000
  storage_type            = "io1"
}

resource "aws_db_instance" "test" {
  allocated_storage   = %[3]d
  identifier          = %[4]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
  iops                = %[5]d
  storage_type        = "io1"
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, allocatedStorage, rName, iops)
}

func testAccInstanceConfig_ReplicateSourceDB_allowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allow_major_version_upgrade = %[2]t
  identifier                  = %[1]q
  instance_class              = aws_db_instance.source.instance_class
  replicate_source_db         = aws_db_instance.source.identifier
  skip_final_snapshot         = true
}
`, rName, allowMajorVersionUpgrade))
}

func testAccInstanceConfig_ReplicateSourceDB_autoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  auto_minor_version_upgrade = %[2]t
  identifier                 = %[1]q
  instance_class             = aws_db_instance.source.instance_class
  replicate_source_db        = aws_db_instance.source.identifier
  skip_final_snapshot        = true
}
`, rName, autoMinorVersionUpgrade))
}

func testAccInstanceConfig_ReplicateSourceDB_availabilityZone(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  availability_zone   = data.aws_availability_zones.available.names[0]
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_backupRetentionPeriod(rName string, backupRetentionPeriod int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_retention_period = %[2]d
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  replicate_source_db     = aws_db_instance.source.identifier
  skip_final_snapshot     = true
}
`, rName, backupRetentionPeriod))
}

// We provide maintenance_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccInstanceConfig_ReplicateSourceDB_backupWindow(rName, backupWindow, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccInstanceConfig_ReplicateSourceDB_dbSubnetGroupName(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alternate" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.alternate.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider = "awsalternate"

  name       = %[1]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

data "aws_rds_engine_version" "default" {
  provider = "awsalternate"
  engine   = %[2]q
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awsalternate"

  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[3]s]
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.alternate.name
  engine                  = data.aws_rds_engine_version.default.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name = aws_db_subnet_group.test.name
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  replicate_source_db  = aws_db_instance.source.arn
  skip_final_snapshot  = true
}
`, rName, tfrds.InstanceEngineMySQL, mainInstanceClasses))
}

// When testing needs to distinguish a second region and second account in the same region
// e.g. cross-region functionality with RAM shared subnets
func testAccAlternateAccountAndAlternateRegionProviderConfig() string {
	// lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternateaccountalternateregion" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}

provider "awsalternateaccountsameregion" {
  access_key = %[1]q
  profile    = %[2]q
  secret_key = %[4]q
}

provider "awssameaccountalternateregion" {
  region = %[3]q
}
`, os.Getenv(envvar.AlternateAccessKeyId), os.Getenv(envvar.AlternateProfile), acctest.AlternateRegion(), os.Getenv(envvar.AlternateSecretAccessKey))
}

func testAccInstanceConfig_ReplicateSourceDB_DBSubnetGroupName_ramShared(rName string) string {
	return acctest.ConfigCompose(testAccAlternateAccountAndAlternateRegionProviderConfig(), fmt.Sprintf(`
data "aws_availability_zones" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_availability_zones" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "sameaccountalternateregion" {
  count    = 2
  provider = "awssameaccountalternateregion"

  availability_zone = data.aws_availability_zones.sameaccountalternateregion.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.sameaccountalternateregion.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alternateaccountsameregion" {
  count    = 2
  provider = "awsalternateaccountsameregion"

  availability_zone = data.aws_availability_zones.alternateaccountsameregion.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.alternateaccountsameregion.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  name = %[1]q
}

resource "aws_ram_principal_association" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.alternateaccountsameregion.arn
}

resource "aws_ram_resource_association" "alternateaccountsameregion" {
  count    = 2
  provider = "awsalternateaccountsameregion"

  resource_arn       = aws_subnet.alternateaccountsameregion[count.index].arn
  resource_share_arn = aws_ram_resource_share.alternateaccountsameregion.id
}

resource "aws_db_subnet_group" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  name       = %[1]q
  subnet_ids = aws_subnet.sameaccountalternateregion[*].id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.alternateaccountsameregion, aws_ram_resource_association.alternateaccountsameregion]

  name       = %[1]q
  subnet_ids = aws_subnet.alternateaccountsameregion[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.alternateaccountsameregion, aws_ram_resource_association.alternateaccountsameregion]

  name   = %[1]q
  vpc_id = aws_vpc.alternateaccountsameregion.id
}

data "aws_rds_engine_version" "default" {
  provider = "awssameaccountalternateregion"
  engine   = %[2]q
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awssameaccountalternateregion"

  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[3]s]
}

resource "aws_db_instance" "source" {
  provider = "awssameaccountalternateregion"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.sameaccountalternateregion.name
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.arn
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName, tfrds.InstanceEngineMySQL, mainInstanceClasses))
}

func testAccInstanceConfig_ReplicateSourceDB_DBSubnetGroupName_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "alternate" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.alternate.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider = "awsalternate"

  name       = %[1]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

data "aws_rds_engine_version" "default" {
  provider = "awsalternate"
  engine   = %[2]q
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awsalternate"

  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[3]s]
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.alternate.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.arn
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName, tfrds.InstanceEngineMySQL, mainInstanceClasses))
}

func testAccInstanceConfig_ReplicateSourceDB_deletionProtection(rName string, deletionProtection bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  deletion_protection = %[2]t
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName, deletionProtection))
}

func testAccInstanceConfig_ReplicateSourceDB_iamDatabaseAuthenticationEnabled(rName string, iamDatabaseAuthenticationEnabled bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  iam_database_authentication_enabled = %[2]t
  identifier                          = %[1]q
  instance_class                      = aws_db_instance.source.instance_class
  replicate_source_db                 = aws_db_instance.source.identifier
  skip_final_snapshot                 = true
}
`, rName, iamDatabaseAuthenticationEnabled))
}

// We provide backup_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccInstanceConfig_ReplicateSourceDB_maintenanceWindow(rName, backupWindow, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccInstanceConfig_ReplicateSourceDB_maxAllocatedStorage(rName string, maxAllocatedStorage int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allocated_storage     = aws_db_instance.source.allocated_storage
  identifier            = %[1]q
  instance_class        = aws_db_instance.source.instance_class
  max_allocated_storage = %[2]d
  replicate_source_db   = aws_db_instance.source.identifier
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccInstanceConfig_ReplicateSourceDB_monitoring(rName string, monitoringInterval int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		testAccInstanceConfig_baseMonitoringRole(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName, monitoringInterval))
}

func testAccInstanceConfig_ReplicateSourceDB_monitoring_sourceOnly(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_multiAZ(rName string, multiAz bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  multi_az            = %[2]t
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName, multiAz))
}

func testAccInstanceConfig_ReplicateSourceDB_networkType(rName string, networkType string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  network_type        = %[2]q
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName, networkType))
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_sameSetOnBoth(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  parameter_group_name    = aws_db_parameter_group.test.name
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.name
  replicate_source_db  = aws_db_instance.source.identifier
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_differentSetOnBoth(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.name
  replicate_source_db  = aws_db_instance.source.identifier
  skip_final_snapshot  = true
}

resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  parameter_group_name    = aws_db_parameter_group.source.name
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_parameter_group" "source" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = "%[1]s-source"
  parameter {
    name  = "sync_binlog"
    value = 0
  }
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_replicaCopiesValue(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  parameter_group_name    = aws_db_parameter_group.test.name
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupName_setOnReplica(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.name
  replicate_source_db  = aws_db_instance.source.identifier
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_port(rName string, port int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  port                = %[2]d
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, rName, port))
}

func testAccInstanceConfig_ReplicateSourceDB_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.identifier
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_caCertificateID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  ca_cert_identifier      = data.aws_rds_certificate.latest.id
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.identifier
  ca_cert_identifier  = data.aws_rds_certificate.latest.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_characterSet_Source(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassOracleEnterprise(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier          = %[1]q
  replicate_source_db = aws_db_instance.source.identifier
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  apply_immediately   = true
}

resource "aws_db_instance" "source" {
  identifier              = "%[1]s-source"
  allocated_storage       = 20
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  db_name                 = "MAINDB"
  username                = "oadmin"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  apply_immediately       = true
  backup_retention_period = 1

  character_set_name = "WE8ISO8859P15"
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_characterSet_Replica(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassOracleEnterprise(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier          = %[1]q
  replicate_source_db = aws_db_instance.source.identifier
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  apply_immediately   = true

  character_set_name = "NE8ISO8859P10"
}

resource "aws_db_instance" "source" {
  identifier              = "%[1]s-source"
  allocated_storage       = 20
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  db_name                 = "MAINDB"
  username                = "oadmin"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  apply_immediately       = true
  backup_retention_period = 1

  character_set_name = "WE8ISO8859P15"
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_replicaMode(rName, replicaMode string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "bring-your-own-license"
  storage_type               = "gp2"
  read_replica_capable       = true
  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "source" {
  identifier              = "%[3]s-source"
  allocated_storage       = 20
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[3]q
  instance_class      = aws_db_instance.source.instance_class
  replica_mode        = %[4]q
  replicate_source_db = aws_db_instance.source.identifier
  skip_final_snapshot = true
}
`, tfrds.InstanceEngineOracleEnterprise, strings.Replace(mainInstanceClasses, "db.t3.small", "frodo", 1), rName, replicaMode)
}

func testAccInstanceConfig_ReplicateSourceDB_ParameterGroupTwoStep_setup(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "bring-your-own-license"
  storage_type               = "gp2"
  read_replica_capable       = true
  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "source" {
  identifier              = "%[3]s-source"
  allocated_storage       = 20
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  db_name                 = "MAINDB"
  username                = "oadmin"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  apply_immediately       = true
  backup_retention_period = 3

  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  character_set_name   = "AL32UTF8"
  timeouts {
    update = "120m"
  }
  ca_cert_identifier = "rds-ca-2019"
}
`, tfrds.InstanceEngineOracleEnterprise, strings.Replace(mainInstanceClasses, "db.t3.small", "frodo", 1), rName)
}

func testAccInstanceConfig_ReplicateSourceDB_parameterGroupTwoStep(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_ReplicateSourceDB_ParameterGroupTwoStep_setup(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier          = %[1]q
  replicate_source_db = aws_db_instance.source.identifier

  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  apply_immediately   = true

  parameter_group_name = aws_db_parameter_group.test.name
  ca_cert_identifier   = "rds-ca-2019"

  timeouts {
    update = "120m"
  }
}

resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q
}
`, rName))
}

func testAccInstanceConfig_ReplicateSourceDB_CrossRegion_ParameterGroupName_equivalent(rName string) string {
	parameters := `
parameter {
  # "max_string_size" cannot be changed after creation
  name         = "max_string_size"
  value        = "EXTENDED"
  apply_method = "immediate"
}
`
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  provider = "aws"

  identifier           = %[1]q
  replicate_source_db  = aws_db_instance.source.arn
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot  = true
  apply_immediately    = true
  parameter_group_name = aws_db_parameter_group.test.name
}

resource "aws_db_parameter_group" "test" {
  provider = "aws"

  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  %[4]s
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  identifier              = "%[1]s-source"
  allocated_storage       = 20
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  db_name                 = "MAINDB"
  username                = "oadmin"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  apply_immediately       = true
  backup_retention_period = 3
  parameter_group_name    = aws_db_parameter_group.source.name
}

resource "aws_db_parameter_group" "source" {
  provider = "awsalternate"

  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = "%[1]s-source"

  %[4]s
}

data "aws_rds_engine_version" "default" {
  engine = %[2]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "bring-your-own-license"
  storage_type               = "gp2"
  read_replica_capable       = true
  preferred_instance_classes = [%[3]s]
}
`, rName, tfrds.InstanceEngineOracleEnterprise, strings.Replace(mainInstanceClasses, "db.t3.small", "frodo", 1), parameters))
}

func testAccInstanceConfig_ReplicateSourceDB_CrossRegion_ParameterGroupName_postgres(rName string) string {
	parameters := `
parameter {
  name         = "client_encoding"
  value        = "UTF8"
  apply_method = "pending-reboot"
}
`
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccInstanceConfig_orderableClassPostgres(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  provider = "aws"

  identifier           = %[1]q
  replicate_source_db  = aws_db_instance.source.arn
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot  = true
  apply_immediately    = true
  parameter_group_name = aws_db_parameter_group.test.name
}

resource "aws_db_parameter_group" "test" {
  provider = "aws"

  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  %[2]s
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  identifier              = "%[1]s-source"
  allocated_storage       = 20
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  db_name                 = "MAINDB"
  username                = "oadmin"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  apply_immediately       = true
  backup_retention_period = 3
  parameter_group_name    = aws_db_parameter_group.source.name
}

resource "aws_db_parameter_group" "source" {
  provider = "awsalternate"

  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = "%[1]s-source"

  %[2]s
}
`, rName, parameters))
}

func testAccInstanceConfig_ReplicateSourceDB_CrossRegion_CharacterSet(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccInstanceConfig_orderableClassOracleEnterprise(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  provider = "aws"

  identifier          = %[1]q
  replicate_source_db = aws_db_instance.source.arn
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  apply_immediately   = true
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  identifier              = "%[1]s-source"
  allocated_storage       = 20
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type            = data.aws_rds_orderable_db_instance.test.storage_type
  db_name                 = "MAINDB"
  username                = "oadmin"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  apply_immediately       = true
  backup_retention_period = 1

  character_set_name = "WE8ISO8859P15"
}
`, rName))
}

func testAccInstanceConfig_baseMonitoringRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}
`, rName)
}

func testAccInstanceConfig_snapshotID(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_snapshotID_ManageMasterPasswordKMSKey(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_kms_key" "example" {
  description = "Terraform acc test %[1]s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
 POLICY

}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier                    = %[1]q
  instance_class                = aws_db_instance.source.instance_class
  snapshot_identifier           = aws_db_snapshot.test.id
  skip_final_snapshot           = true
  manage_master_user_password   = true
  master_user_secret_kms_key_id = aws_kms_key.example.arn
}
`, rName))
}

func testAccInstanceConfig_SnapshotIdentifier_namePrefix(identifierPrefix, sourceName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier_prefix   = %[2]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, sourceName, identifierPrefix))
}

func testAccInstanceConfig_SnapshotIdentifier_nameGenerated(sourceName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, sourceName))
}

func testAccInstanceConfig_SnapshotID_associationRemoved(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_allocatedStorage(rName string, allocatedStorage int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage   = %[2]d
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, allocatedStorage))
}

func testAccInstanceConfig_SnapshotID_ioStorage(rName string, sType string, iops int) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                = %[1]q
  engine_latest_version = true
  license_model         = "general-public-license"
  storage_type          = %[4]q

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 200
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[3]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[3]q
}

resource "aws_db_instance" "test" {
  identifier          = %[3]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
  allocated_storage   = 200
  iops                = %[5]d
  storage_type        = data.aws_rds_orderable_db_instance.test.storage_type
}
`, tfrds.InstanceEngineMariaDB, mainInstanceClasses, rName, sType, iops)
}

func testAccInstanceConfig_SnapshotID_allowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                  = %[1]q
  latest                  = true
  preferred_major_targets = [data.aws_rds_engine_version.upgrade.version_actual]
}

data "aws_rds_engine_version" "upgrade" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "postgres13" {
  engine         = %[1]q
  engine_version = data.aws_rds_engine_version.test.version_actual
  license_model  = "postgresql-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.postgres13.engine
  engine_version      = data.aws_rds_orderable_db_instance.postgres13.engine_version
  identifier          = "%[3]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.postgres13.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[3]q
}

data "aws_rds_orderable_db_instance" "postgres14" {
  engine         = %[1]q
  engine_version = data.aws_rds_engine_version.upgrade.version_actual
  license_model  = "postgresql-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "test" {
  allow_major_version_upgrade = %[4]t
  engine                      = data.aws_rds_orderable_db_instance.postgres14.engine
  engine_version              = data.aws_rds_orderable_db_instance.postgres14.engine_version
  identifier                  = %[3]q
  instance_class              = aws_db_instance.source.instance_class
  snapshot_identifier         = aws_db_snapshot.test.id
  skip_final_snapshot         = true
}
`, tfrds.InstanceEnginePostgres, mainInstanceClasses, rName, allowMajorVersionUpgrade)
}

func testAccInstanceConfig_SnapshotID_autoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  auto_minor_version_upgrade = %[2]t
  identifier                 = %[1]q
  instance_class             = aws_db_instance.source.instance_class
  snapshot_identifier        = aws_db_snapshot.test.id
  skip_final_snapshot        = true
}
`, rName, autoMinorVersionUpgrade))
}

func testAccInstanceConfig_SnapshotID_availabilityZone(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  availability_zone   = data.aws_availability_zones.available.names[0]
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_backupRetentionPeriod(rName string, backupRetentionPeriod int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_retention_period = %[2]d
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName, backupRetentionPeriod))
}

func testAccInstanceConfig_SnapshotID_BackupRetentionPeriod_unset(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_retention_period = 0
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName))
}

// We provide maintenance_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccInstanceConfig_SnapshotID_backupWindow(rName, backupWindow, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccInstanceConfig_SnapshotID_dbSubnetGroupName(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name = aws_db_subnet_group.test.name
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  snapshot_identifier  = aws_db_snapshot.test.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_DBSubnetGroupName_ramShared(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_association" "test" {
  count    = 2
  provider = "awsalternate"

  resource_arn       = aws_subnet.test[count.index].arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_DBSubnetGroupName_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_deletionProtection(rName string, deletionProtection bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  deletion_protection = %[2]t
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, deletionProtection))
}

func testAccInstanceConfig_SnapshotID_iamDatabaseAuthenticationEnabled(rName string, iamDatabaseAuthenticationEnabled bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMySQL(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  iam_database_authentication_enabled = %[2]t
  identifier                          = %[1]q
  instance_class                      = aws_db_instance.source.instance_class
  snapshot_identifier                 = aws_db_snapshot.test.id
  skip_final_snapshot                 = true
}
`, rName, iamDatabaseAuthenticationEnabled))
}

// We provide backup_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccInstanceConfig_SnapshotID_maintenanceWindow(rName, backupWindow, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccInstanceConfig_SnapshotID_maxAllocatedStorage(rName string, maxAllocatedStorage int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage     = aws_db_instance.source.allocated_storage
  identifier            = %[1]q
  instance_class        = aws_db_instance.source.instance_class
  max_allocated_storage = %[2]d
  snapshot_identifier   = aws_db_snapshot.test.id
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccInstanceConfig_SnapshotID_monitoring(rName string, monitoringInterval int) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, monitoringInterval))
}

func testAccInstanceConfig_SnapshotID_multiAZ(rName string, multiAz bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  multi_az            = %[2]t
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, multiAz))
}

func testAccInstanceConfig_SnapshotID_MultiAZ_sqlServer(rName string, multiAz bool) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassSQLServerSe(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 20
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  # InvalidParameterValue: Mirroring cannot be applied to instances with backup retention set to zero.
  backup_retention_period = 1
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  multi_az                = %[2]t
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName, multiAz))
}

func testAccInstanceConfig_SnapshotID_parameterGroupName(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.name
  snapshot_identifier  = aws_db_snapshot.test.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_port(rName string, port int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  port                = %[2]d
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, port))
}

func testAccInstanceConfig_SnapshotID_tags(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true

  tags = {
    key1 = "value-old"
  }
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true

  tags = {
    key1 = "value1"
  }
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_tagsRemove(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true

  tags = {
    key1 = "value1"
  }
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true

  tags = {}
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccInstanceConfig_SnapshotID_VPCSecurityGroupIDs_tags(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]

  tags = {
    key1 = "value1"
  }
}
`, rName))
}

func testAccInstanceConfig_performanceInsightsDisabled(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage       = 5
  backup_retention_period = 0
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = %[3]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "mydb"
  password                = "mustbeeightcharaters"
  skip_final_snapshot     = true
  username                = "foo"
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName)
}

func testAccInstanceConfig_performanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %[3]q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                               = "mydb"
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_retention_period = 7
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName)
}

func testAccInstanceConfig_performanceInsightsKMSKeyIdDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage       = 5
  backup_retention_period = 0
  db_name                 = "mydb"
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = %[3]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "mustbeeightcharaters"
  skip_final_snapshot     = true
  username                = "foo"
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName)
}

func testAccInstanceConfig_performanceInsightsKMSKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  db_name                               = "mydb"
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %[3]q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName)
}

func testAccInstanceConfig_performanceInsightsRetentionPeriod(rName string, performanceInsightsRetentionPeriod int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %[3]q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                               = "mydb"
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_retention_period = %[4]d
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName, performanceInsightsRetentionPeriod)
}

func testAccInstanceConfig_ReplicateSourceDB_performanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  description = "Terraform acc test"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = [%[2]s]
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = "%[3]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "mustbeeightcharaters"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier                            = %[3]q
  instance_class                        = aws_db_instance.source.instance_class
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  replicate_source_db                   = aws_db_instance.source.identifier
  skip_final_snapshot                   = true
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName)
}

func testAccInstanceConfig_SnapshotID_performanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  description = "Terraform acc test"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = [%[2]s]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = "%[3]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.identifier
  db_snapshot_identifier = %[3]q
}

resource "aws_db_instance" "test" {
  identifier                            = %[3]q
  instance_class                        = aws_db_instance.source.instance_class
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  snapshot_identifier                   = aws_db_snapshot.test.id
  skip_final_snapshot                   = true
}
`, tfrds.InstanceEngineMySQL, mainInstanceClasses, rName)
}

func testAccInstanceConfig_dedicatedLogVolumeEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassPostgres(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  # Dedicated log volumes do not support PG 16 instances.
  engine              = "postgres"
  engine_version      = "15.6"
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
  apply_immediately   = true

  # Minimum amounts required to qualify for IOPS / DedicatedLogVolume
  allocated_storage = 100
  storage_type      = "io1"
  iops              = 1000

  dedicated_log_volume = %[2]t
}
`, rName, enabled))
}

func testAccInstanceConfig_noDeleteAutomatedBackups(rName string) string {
	return acctest.ConfigCompose(testAccInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true

  backup_retention_period  = 1
  delete_automated_backups = false
}
`, rName))
}

func testAccInstanceConfig_baseOutpost(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClass(tfrds.InstanceEngineMySQL, "general-public-license", "standard"),
		fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.128.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.128.1.0/24"
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  vpc_id            = aws_vpc.test.id
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.id]
}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = data.aws_outposts_outpost.test.arn
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_Outpost_coIPEnabled(rName string, coipEnabled bool, backupRetentionPeriod int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseOutpost(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier                = %[1]q
  allocated_storage         = 20
  backup_retention_period   = %[3]d
  engine                    = data.aws_rds_orderable_db_instance.test.engine
  engine_version            = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class            = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                   = "test"
  parameter_group_name      = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot       = true
  password                  = "avoid-plaintext-passwords"
  username                  = "tfacctest"
  db_subnet_group_name      = aws_db_subnet_group.test.name
  storage_encrypted         = true
  customer_owned_ip_enabled = %[2]t
}
`, rName, coipEnabled, backupRetentionPeriod))
}

func testAccInstanceConfig_Outposts_coIPRestorePointInTime(rName string, sourceCoipEnabled bool, targetCoipEnabled bool) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_Outpost_coIPEnabled(rName, sourceCoipEnabled, 1),
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = "%[1]s-restore"
  instance_class = aws_db_instance.test.instance_class
  restore_to_point_in_time {
    source_db_instance_identifier = aws_db_instance.test.identifier
    use_latest_restorable_time    = true
  }
  storage_encrypted         = true
  skip_final_snapshot       = true
  db_subnet_group_name      = aws_db_instance.test.db_subnet_group_name
  customer_owned_ip_enabled = %[2]t
}
`, rName, targetCoipEnabled))
}

func testAccInstanceConfig_Outposts_coIPSnapshotID(rName string, sourceCoipEnabled bool, targetCoipEnabled bool) string {
	return acctest.ConfigCompose(testAccInstanceConfig_Outpost_coIPEnabled(rName, sourceCoipEnabled, 1), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "restore" {
  customer_owned_ip_enabled = %[2]t
  db_subnet_group_name      = aws_db_subnet_group.foo.name
  storage_encrypted         = true
  identifier                = "%[1]s-restore"
  instance_class            = aws_db_instance.test.instance_class
  snapshot_identifier       = aws_db_snapshot.test.id
  skip_final_snapshot       = true
}
`, rName, targetCoipEnabled))
}

func testAccInstanceConfig_Outposts_backupTarget(rName string, backupTarget string, backupRetentionPeriod int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_baseOutpost(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 20
  backup_retention_period = %[3]d
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  db_subnet_group_name    = aws_db_subnet_group.test.name
  storage_encrypted       = true
  backup_target           = %[2]q
}
`, rName, backupTarget, backupRetentionPeriod))
}

func testAccInstanceConfig_license(rName, license string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClass(tfrds.InstanceEngineOracleStandard2, license, "standard"),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  apply_immediately   = true
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName))
}

func testAccInstanceConfig_customIAMInstanceProfile(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassCustomSQLServerWeb(),
		fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name          = %[1]q
  capabilities  = ["CAPABILITY_NAMED_IAM"]
  template_body = file("test-fixtures/custom-sql-cloudformation.json")
}

resource "aws_db_instance" "test" {
  allocated_storage           = 20
  auto_minor_version_upgrade  = false
  custom_iam_instance_profile = aws_cloudformation_stack.test.outputs["RDSCustomSQLServerInstanceProfile"]
  engine                      = data.aws_rds_engine_version.default.engine
  identifier                  = %[1]q
  instance_class              = data.aws_rds_orderable_db_instance.test.instance_class
  kms_key_id                  = aws_cloudformation_stack.test.outputs["RDSCustomSQLServerKMSKey"]
  password                    = "avoid-plaintext-passwords"
  username                    = "tfacctest"
  skip_final_snapshot         = true
  storage_encrypted           = true
  vpc_security_group_ids      = [aws_cloudformation_stack.test.outputs["RDSCustomSecurityGroup"]]
  db_subnet_group_name        = aws_cloudformation_stack.test.outputs["DBSubnetGroup"]
}
`, rName))
}

func testAccInstanceConfig_BlueGreenDeployment_engineVersion(rName string, update bool) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${local.engine_version.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  blue_green_update {
    enabled = true
  }
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = local.engine_version.engine
  engine_version = local.engine_version.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[2]s]
}

data "aws_rds_engine_version" "initial" {
  engine                    = %[3]q
  latest                    = true
  preferred_upgrade_targets = [data.aws_rds_engine_version.update.version_actual]
}

data "aws_rds_engine_version" "update" {
  engine = %[3]q
}

locals {
  engine_version = %[4]t ? data.aws_rds_engine_version.update : data.aws_rds_engine_version.initial
}
`, rName, mainInstanceClasses, tfrds.InstanceEngineMySQL, update))
}

func testAccInstanceConfig_BlueGreenDeployment_pre(rName string, oddClasses bool) string {
	var halfClasses []string
	start := 0
	if oddClasses {
		start = 1
	}
	for i := start; i < len(instanceClassesSlice); i += 2 {
		halfClasses = append(halfClasses, instanceClassesSlice[i])
	}
	halfMainInstClass := strings.Join(halfClasses, ", ")

	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[4]s]
}

resource "aws_db_instance" "test" {
  identifier              = %[5]q
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, tfrds.InstanceEngineMySQL, "general-public-license", "standard", halfMainInstClass, rName)
}

func testAccInstanceConfig_BlueGreenDeployment_parameterGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = aws_db_parameter_group.test.name
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  blue_green_update {
    enabled = true
  }
}

resource "aws_db_parameter_group" "test" {
  family = data.aws_rds_engine_version.default.parameter_group_family
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}
`, rName))
}

func testAccInstanceConfig_BlueGreenDeployment_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  tags = {
    %[2]q = %[3]q
  }

  blue_green_update {
    enabled = true
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccInstanceConfig_BlueGreenDeployment_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
data "aws_db_parameter_group" "test" {
  name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
}

resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  blue_green_update {
    enabled = true
  }
}
`, rName))
}

func testAccInstanceConfig_BlueGreenDeployment_updateableInstanceClass(rName string, oddClasses bool) string {
	var halfClasses []string
	start := 0
	if oddClasses {
		start = 1
	}
	for i := start; i < len(instanceClassesSlice); i += 2 {
		halfClasses = append(halfClasses, instanceClassesSlice[i])
	}
	halfMainInstClass := strings.Join(halfClasses, ", ")

	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[4]s]
}

data "aws_db_parameter_group" "test" {
  name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
}

resource "aws_db_instance" "test" {
  identifier              = %[5]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = data.aws_db_parameter_group.test.name
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  blue_green_update {
    enabled = true
  }
}
`, tfrds.InstanceEngineMySQL, "general-public-license", "standard", halfMainInstClass, rName)
}

func testAccInstanceConfig_BlueGreenDeployment_prePromote(rName string) string {
	var e []string
	for i := 0; i < len(instanceClassesSlice); i += 2 {
		e = append(e, instanceClassesSlice[i])
	}
	evenClasses := strings.Join(e, ", ")

	var o []string
	for i := 1; i < len(instanceClassesSlice); i += 2 {
		o = append(o, instanceClassesSlice[i])
	}
	oddClasses := strings.Join(o, ", ")

	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[4]s]
}

data "aws_rds_orderable_db_instance" "update" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[5]s]
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[6]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_retention_period = 1
  identifier              = %[6]q
  instance_class          = aws_db_instance.source.instance_class
  replicate_source_db     = aws_db_instance.source.identifier
  skip_final_snapshot     = true
}
`, tfrds.InstanceEngineMySQL, "general-public-license", "standard", oddClasses, evenClasses, rName)
}

func testAccInstanceConfig_BlueGreenDeployment_promote(rName string) string {
	var e []string
	for i := 0; i < len(instanceClassesSlice); i += 2 {
		e = append(e, instanceClassesSlice[i])
	}
	evenClasses := strings.Join(e, ", ")

	var o []string
	for i := 1; i < len(instanceClassesSlice); i += 2 {
		o = append(o, instanceClassesSlice[i])
	}
	oddClasses := strings.Join(o, ", ")

	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[4]s]
}

data "aws_rds_orderable_db_instance" "update" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[5]s]
}

resource "aws_db_instance" "source" {
  identifier              = "%[6]s-source"
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[6]q
  instance_class      = data.aws_rds_orderable_db_instance.update.instance_class
  skip_final_snapshot = true

  blue_green_update {
    enabled = true
  }
}
`, tfrds.InstanceEngineMySQL, "general-public-license", "standard", oddClasses, evenClasses, rName)
}

func testAccInstanceConfig_BlueGreenDeployment_deletionProtection(rName string, deletionProtection bool, oddClasses bool) string {
	var halfClasses []string
	start := 0
	if oddClasses {
		start = 1
	}
	for i := start; i < len(instanceClassesSlice); i += 2 {
		halfClasses = append(halfClasses, instanceClassesSlice[i])
	}
	halfMainInstClass := strings.Join(halfClasses, ", ")

	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = %[2]q
  storage_type   = %[3]q

  preferred_instance_classes = [%[4]s]
}

resource "aws_db_instance" "test" {
  identifier              = %[5]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"

  blue_green_update {
    enabled = true
  }

  deletion_protection = %[6]t
}
`, tfrds.InstanceEngineMySQL, "general-public-license", "standard", halfMainInstClass, rName, deletionProtection)
}

func testAccInstanceConfig_BlueGreenDeployment_password(rName, password string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQL(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot     = true
  password                = %[2]q
  username                = "tfacctest"

  blue_green_update {
    enabled = true
  }
}
`, rName, password))
}

func testAccInstanceConfig_engineVersion(rName string, update bool) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  parameter_group_name    = "default.${local.engine_version.parameter_group_family}"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = local.engine_version.engine
  engine_version = local.engine_version.version
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = [%[2]s]
}

data "aws_rds_engine_version" "initial" {
  engine                    = %[3]q
  latest                    = true
  preferred_upgrade_targets = [data.aws_rds_engine_version.update.version_actual]
}

data "aws_rds_engine_version" "update" {
  engine = %[3]q
}

locals {
  engine_version = %[4]t ? data.aws_rds_engine_version.update : data.aws_rds_engine_version.initial
}
`, rName, mainInstanceClasses, tfrds.InstanceEngineMySQL, update))
}

func testAccInstanceConfig_Storage_gp3(rName string, orderableClassConfig func() string, allocatedStorage int) string {
	return acctest.ConfigCompose(
		orderableClassConfig(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = data.aws_rds_engine_version.default.engine == "%[2]s" ? null : "test" # using %[2]q breaks linter
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true

  storage_type      = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage = %[3]d
}
`, rName, tfrds.InstanceEngineSQLServerExpress, allocatedStorage))
}

func testAccInstanceConfig_Storage_throughput(rName string, iops, throughput int) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMySQLGP3(),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier           = %[1]q
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true

  storage_type      = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage = 400

  iops               = %[2]d
  storage_throughput = %[3]d
}
`, rName, iops, throughput))
}

func testAccInstanceConfig_Storage_throughputSSE(rName string, iops, throughput int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "license-included"
  storage_type   = "gp3"

  preferred_instance_classes = [%[2]s]
}

resource "aws_db_instance" "test" {
  allocated_storage    = 400
  apply_immediately    = true
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  identifier           = %[3]q
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  iops                 = %[4]d
  license_model        = data.aws_rds_orderable_db_instance.test.license_model
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password             = "avoid-plaintext-passwords"
  skip_final_snapshot  = true
  storage_throughput   = %[5]d
  storage_type         = data.aws_rds_orderable_db_instance.test.storage_type
  username             = "tfacctest"
}
`, tfrds.InstanceEngineSQLServerStandard, mainInstanceClasses, rName, iops, throughput)
}

func testAccInstanceConfig_Storage_typePostgres(rName string, storageType string, allocatedStorage int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  storage_type   = %[2]q

  preferred_instance_classes = [%[3]s]
}

resource "aws_db_instance" "test" {
  identifier           = %[4]q
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true

  storage_type      = %[2]q
  allocated_storage = %[5]d
}
`, tfrds.InstanceEnginePostgres, storageType, mainInstanceClasses, rName, allocatedStorage)
}
