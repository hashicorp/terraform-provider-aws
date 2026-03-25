// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSClusterInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rds", regexache.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "custom_iam_instance_profile", ""),
					resource.TestCheckResourceAttrSet(resourceName, "dbi_resource_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "aurora-mysql"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPreferredMaintenanceWindow),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfrds.ResourceClusterInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSClusterInstance_identifierGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_identifierGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrIdentifier, "tf-"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", "tf-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_identifierPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_identifierPrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
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
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_isAlreadyBeingDeleted(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					// Get Database Instance into deleting state
					conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)
					input := &rds.DeleteDBInstanceInput{
						DBInstanceIdentifier: aws.String(rName),
						SkipFinalSnapshot:    aws.Bool(true),
					}
					_, err := conn.DeleteDBInstance(ctx, input)
					if err != nil {
						t.Fatalf("error deleting Database Instance: %s", err)
					}
				},
				Config:  testAccClusterInstanceConfig_basic(rName),
				Destroy: true,
			},
		},
	})
}

func TestAccRDSClusterInstance_az(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_az(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, availabilityZonesDataSourceName, "names.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_publiclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_publiclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_publiclyAccessible(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_publiclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_copyTagsToSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_copyTagsToSnapshot(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_copyTagsToSnapshot(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_caCertificateIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_instance.test"
	certificateDataSourceName := "data.aws_rds_certificate.latest"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_caCertificateID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", certificateDataSourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_monitoringInterval(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_monitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_monitoringInterval(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "60"),
				),
			},
			{
				Config: testAccClusterInstanceConfig_monitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
			{
				Config: testAccClusterInstanceConfig_monitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_MonitoringRoleARN_enabledToDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_monitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_MonitoringRoleARN_enabledToRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_monitoringRoleARNRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_MonitoringRoleARN_removedToEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_monitoringRoleARNRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_monitoringRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsEnabled_auroraMySQL1(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	engine := "aurora-mysql"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, engine) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_performanceInsightsEnabledAuroraMySQL1(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsEnabled_auroraPostgresql(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	engine := "aurora-postgresql"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, engine) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_performanceInsightsEnabledAuroraPostgresql(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyID_auroraMySQL1(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	engine := "aurora-mysql"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, engine) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_performanceInsightsKMSKeyIDAuroraMySQL1(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
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
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyIDAuroraMySQL1_defaultKeyToCustomKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	engine := "aurora-mysql"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, engine) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_performanceInsightsEnabledAuroraMySQL1(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config:      testAccClusterInstanceConfig_performanceInsightsKMSKeyIDAuroraMySQL1(rName, engine),
				ExpectError: regexache.MustCompile(`InvalidParameterCombination: You .* change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccRDSClusterInstance_performanceInsightsRetentionPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, "aurora-mysql")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_performanceInsightsRetentionPeriod(rName, 731),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
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
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccClusterInstanceConfig_performanceInsightsRetentionPeriod(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
			{
				Config: testAccClusterInstanceConfig_performanceInsightsRetentionPeriod(rName, 155),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "155"),
				),
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyID_auroraPostgresql(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	engine := "aurora-postgresql"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, engine) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_performanceInsightsKMSKeyIDAuroraPostgresql(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
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
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRDSClusterInstance_PerformanceInsightsKMSKeyIDAuroraPostgresql_defaultKeyToCustomKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	engine := "aurora-postgresql"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPerformanceInsightsDefaultVersionPreCheck(ctx, t, engine) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_performanceInsightsEnabledAuroraPostgresql(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					names.AttrForceDestroy,
				},
			},
			{
				Config:      testAccClusterInstanceConfig_performanceInsightsKMSKeyIDAuroraPostgresql(rName, engine),
				ExpectError: regexache.MustCompile(`InvalidParameterCombination: You .* change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccRDSClusterInstance_Replica_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var primaryInstance types.DBInstance
	var replicaInstance types.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	resourceName2 := "aws_rds_cluster_instance.alternate"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// record the initialized providers so that we can use them to
	// check for the cluster in each region
	var providers []*schema.Provider

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckClusterInstanceDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceConfig_replicationSource_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterInstanceExistsWithProvider(ctx, resourceName, &primaryInstance, acctest.RegionProviderFunc(ctx, acctest.Region(), &providers)),
					testAccCheckClusterInstanceExistsWithProvider(ctx, resourceName2, &replicaInstance, acctest.RegionProviderFunc(ctx, acctest.AlternateRegion(), &providers)),
				),
			},
		},
	})
}

func testAccCheckClusterInstanceExistsWithProvider(ctx context.Context, n string, v *types.DBInstance, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := providerF().Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBInstanceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterInstanceDestroyWithProvider(ctx context.Context) acctest.TestCheckWithProviderFunc {
	return func(s *terraform.State, provider *schema.Provider) error {
		conn := provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster_instance" {
				continue
			}

			_, err := tfrds.FindDBInstanceByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Cluster Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPerformanceInsightsDefaultVersionPreCheck(ctx context.Context, t *testing.T, engine string) {
	conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

	input := &rds.DescribeDBEngineVersionsInput{
		DefaultOnly: aws.Bool(true),
		Engine:      aws.String(engine),
	}

	result, err := conn.DescribeDBEngineVersions(ctx, input)
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(result.DBEngineVersions) < 1 {
		t.Fatalf("unexpected PreCheck error, no default version for engine: %s", engine)
	}

	testAccPerformanceInsightsPreCheck(ctx, t, engine, aws.ToString(result.DBEngineVersions[0].EngineVersion))
}

func testAccPerformanceInsightsPreCheck(ctx context.Context, t *testing.T, engine string, engineVersion string) {
	conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

	input := &rds.DescribeOrderableDBInstanceOptionsInput{
		Engine:        aws.String(engine),
		EngineVersion: aws.String(engineVersion), // version cuts response time
	}

	supportsPerformanceInsights := false
	pages := rds.NewDescribeOrderableDBInstanceOptionsPaginator(conn, input)
Pages:
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, "InvalidParameterCombination") {
			t.Skipf("RDS Performance Insights not supported, skipping acceptance test")
		}

		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}

		for _, instanceOption := range page.OrderableDBInstanceOptions {
			if aws.ToBool(instanceOption.SupportsPerformanceInsights) {
				supportsPerformanceInsights = true
				break Pages
			}
		}
	}

	if !supportsPerformanceInsights {
		t.Skipf("RDS Performance Insights not supported, skipping acceptance test")
	}
}

func testAccCheckClusterInstanceExists(ctx context.Context, t *testing.T, n string, v *types.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		output, err := tfrds.FindDBInstanceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterInstanceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster_instance" {
				continue
			}

			_, err := tfrds.FindDBInstanceByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Cluster Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccClusterInstanceConfig_orderableEngineBase(engine string, performanceInsights bool) string {
	if performanceInsights {
		return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}
`, engine)
	}

	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = %[1]q
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}
`, engine)
}

func testAccClusterInstanceConfig_base(rName, engine string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		testAccClusterInstanceConfig_orderableEngineBase(engine, false),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier = %[2]q
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}
`, engine, rName))
}

func testAccClusterInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  identifier              = %[1]q
  engine                  = data.aws_rds_engine_version.default.engine
  cluster_identifier      = aws_rds_cluster.test.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.test.name
  promotion_tier          = "3"
}

resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }
}
`, rName))
}

func testAccClusterInstanceConfig_identifierGenerated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccClusterInstanceConfig_orderableEngineBase("aurora-mysql", false),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  engine             = data.aws_rds_engine_version.default.engine
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName))
}

func testAccClusterInstanceConfig_identifierPrefix(rName, identifierPrefix string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccClusterInstanceConfig_orderableEngineBase("aurora-mysql", false),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  db_subnet_group_name = aws_db_subnet_group.test.name
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  master_password      = "password"
  master_username      = "root"
  skip_final_snapshot  = true
}

resource "aws_rds_cluster_instance" "test" {
  identifier_prefix  = %[2]q
  engine             = data.aws_rds_engine_version.default.engine
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName, identifierPrefix))
}

func testAccClusterInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  identifier              = %[1]q
  engine                  = data.aws_rds_engine_version.default.engine
  cluster_identifier      = aws_rds_cluster.test.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.test.name
  promotion_tier          = "3"

  tags = {
    %[2]q = %[3]q
  }
}

resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  identifier              = %[1]q
  engine                  = data.aws_rds_engine_version.default.engine
  cluster_identifier      = aws_rds_cluster.test.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.test.name
  promotion_tier          = "3"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterInstanceConfig_modified(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  identifier                 = %[1]q
  engine                     = data.aws_rds_engine_version.default.engine
  cluster_identifier         = aws_rds_cluster.test.id
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name    = aws_db_parameter_group.test.name
  auto_minor_version_upgrade = false
  promotion_tier             = "3"
}

resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }
}
`, rName))
}

func testAccClusterInstanceConfig_az(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  identifier              = %[1]q
  engine                  = data.aws_rds_engine_version.default.engine
  cluster_identifier      = aws_rds_cluster.test.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.test.name
  promotion_tier          = "3"
  availability_zone       = data.aws_availability_zones.available.names[0]
}

resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }
}
`, rName))
}

func testAccClusterInstanceConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		testAccClusterInstanceConfig_orderableEngineBase("aurora-mysql", false),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

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
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_rds_cluster" "test" {
  cluster_identifier = %[1]q
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  storage_encrypted   = true
  kms_key_id          = aws_kms_key.test.arn
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  identifier              = %[1]q
  cluster_identifier      = aws_rds_cluster.test.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.test.name
  engine                  = data.aws_rds_engine_version.default.engine
}

resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }
}
`, rName))
}

func testAccClusterInstanceConfig_publiclyAccessible(rName string, publiclyAccessible bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnetsEnableDNSHostnames(rName, 2),
		testAccClusterInstanceConfig_orderableEngineBase("aurora-mysql", false),
		fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  apply_immediately   = true
  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  engine              = aws_rds_cluster.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  publicly_accessible = %[2]t

  depends_on = [aws_internet_gateway.test]
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  database_name        = "mydb"
  master_username      = "foo"
  master_password      = "mustbeeightcharacters"
  skip_final_snapshot  = true
  db_subnet_group_name = aws_db_subnet_group.test.name
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName, publiclyAccessible))
}

func testAccClusterInstanceConfig_copyTagsToSnapshot(rName string, copy bool) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  identifier            = %[1]q
  engine                = data.aws_rds_engine_version.default.engine
  cluster_identifier    = aws_rds_cluster.test.id
  instance_class        = data.aws_rds_orderable_db_instance.test.instance_class
  promotion_tier        = "3"
  copy_tags_to_snapshot = %[2]t
}
`, rName, copy))
}

func testAccClusterInstanceConfig_caCertificateID(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.test.id
  engine             = data.aws_rds_engine_version.default.engine
  identifier         = %[1]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  ca_cert_identifier = data.aws_rds_certificate.latest.id
}
`, rName))
}

func testAccClusterInstanceConfig_monitoringInterval(rName string, monitoringInterval int) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
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
        "Service": "monitoring.rds.amazonaws.com"
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

resource "aws_rds_cluster_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  cluster_identifier  = aws_rds_cluster.test.id
  engine              = data.aws_rds_engine_version.default.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
}
`, rName, monitoringInterval))
}

func testAccClusterInstanceConfig_monitoringRoleARNRemoved(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  engine             = data.aws_rds_engine_version.default.engine
  identifier         = %[1]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}
`, rName))
}

func testAccClusterInstanceConfig_monitoringRoleARN(rName string) string {
	return acctest.ConfigCompose(testAccClusterInstanceConfig_base(rName, "aurora-mysql"), fmt.Sprintf(`
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
        "Service": "monitoring.rds.amazonaws.com"
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

resource "aws_rds_cluster_instance" "test" {
  engine              = data.aws_rds_engine_version.default.engine
  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = 5
  monitoring_role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccClusterInstanceConfig_performanceInsightsEnabledAuroraMySQL1(rName, engine string) string {
	return acctest.ConfigCompose(
		testAccClusterInstanceConfig_orderableEngineBase(engine, true),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = data.aws_rds_engine_version.default.engine
  identifier                   = %[1]q
  instance_class               = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled = true
}
`, rName))
}

func testAccClusterInstanceConfig_replicationSource_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_replicationSource_base(rName),
		fmt.Sprintf(`
resource "aws_rds_cluster" "alternate" {
  provider = "awsalternate"

  cluster_identifier            = "%[1]s-replica"
  db_subnet_group_name          = aws_db_subnet_group.test.name
  engine                        = %[2]q
  kms_key_id                    = aws_kms_key.test.arn
  storage_encrypted             = true
  skip_final_snapshot           = true
  replication_source_identifier = aws_rds_cluster.test.arn
  source_region                 = data.aws_region.current.region

  depends_on = [
    aws_rds_cluster_instance.test,
  ]
}

resource "aws_rds_cluster_instance" "alternate" {
  provider = "awsalternate"

  identifier         = "%[1]s-replica"
  cluster_identifier = aws_rds_cluster.alternate.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  engine             = aws_rds_cluster.alternate.engine
  engine_version     = aws_rds_cluster.alternate.engine_version

  force_destroy = true
}
`, rName, tfrds.ClusterEngineAuroraMySQL))
}

func testAccClusterInstanceConfig_performanceInsightsEnabledAuroraPostgresql(rName, engine string) string {
	return acctest.ConfigCompose(
		testAccClusterInstanceConfig_orderableEngineBase(engine, true),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  identifier                   = %[1]q
  instance_class               = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled = true
}
`, rName))
}

func testAccClusterInstanceConfig_performanceInsightsKMSKeyIDAuroraMySQL1(rName, engine string) string {
	return acctest.ConfigCompose(
		testAccClusterInstanceConfig_orderableEngineBase(engine, true),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName))
}

func testAccClusterInstanceConfig_performanceInsightsKMSKeyIDAuroraPostgresql(rName, engine string) string {
	return acctest.ConfigCompose(
		testAccClusterInstanceConfig_orderableEngineBase(engine, true),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName))
}

func testAccClusterInstanceConfig_performanceInsightsRetentionPeriod(rName string, performanceInsightsRetentionPeriod int) string {
	return acctest.ConfigCompose(
		testAccClusterInstanceConfig_orderableEngineBase("aurora-mysql", true),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier                    = aws_rds_cluster.test.id
  engine                                = aws_rds_cluster.test.engine
  identifier                            = %[1]q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled          = true
  performance_insights_retention_period = %[2]d
}
`, rName, performanceInsightsRetentionPeriod))
}
