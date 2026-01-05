// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftimestreaminfluxdb "github.com/hashicorp/terraform-provider-aws/internal/service/timestreaminfluxdb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamInfluxDBDBCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("timestream-influxdb", regexache.MustCompile(`db-cluster/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("db_storage_type"), tfknownvalue.StringExact(awstypes.DbStorageTypeInfluxIoIncludedT1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deployment_type"), tfknownvalue.StringExact(awstypes.ClusterDeploymentTypeMultiNodeReadReplicas)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("engine_type"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("failover_mode"), tfknownvalue.StringExact(awstypes.FailoverModeAutomatic)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("influx_auth_parameters_secret_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("network_type"), tfknownvalue.StringExact(awstypes.NetworkTypeIpv4)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPort), knownvalue.Int32Exact(8086)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPubliclyAccessible), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reader_endpoint"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftimestreaminfluxdb.ResourceDBCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_dbInstanceType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_dbInstanceType(rName, string(awstypes.DbInstanceTypeDbInfluxMedium)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", string(awstypes.DbInstanceTypeDbInfluxMedium)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
			{
				Config: testAccDBClusterConfig_dbInstanceType(rName, string(awstypes.DbInstanceTypeDbInfluxLarge)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", string(awstypes.DbInstanceTypeDbInfluxLarge)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_logDeliveryConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_logDeliveryConfigurationEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
			{
				Config: testAccDBClusterConfig_logDeliveryConfigurationEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.enabled", acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_networkType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_networkTypeIPV4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "network_type", string(awstypes.NetworkTypeIpv4)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_port(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 timestreaminfluxdb.GetDbClusterOutput
	port1 := "8086"
	port2 := "8087"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_port(rName, port1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, port1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
			{
				Config: testAccDBClusterConfig_port(rName, port2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, port2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_allocatedStorage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	allocatedStorage := "20"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_allocatedStorage(rName, allocatedStorage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrAllocatedStorage, allocatedStorage),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_dbStorageType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_dbStorageType(rName, string(awstypes.DbStorageTypeInfluxIoIncludedT1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "db_storage_type", string(awstypes.DbStorageTypeInfluxIoIncludedT1)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_publiclyAccessible(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_publiclyAccessible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_deploymentType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_deploymentType(rName, string(awstypes.ClusterDeploymentTypeMultiNodeReadReplicas)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.ClusterDeploymentTypeMultiNodeReadReplicas)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_failoverMode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_failoverMode(rName, string(awstypes.FailoverModeAutomatic)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "failover_mode", string(awstypes.FailoverModeAutomatic)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
			{
				Config: testAccDBClusterConfig_failoverMode(rName, string(awstypes.FailoverModeNoFailover)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "failover_mode", string(awstypes.FailoverModeNoFailover)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization"},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_dbParameterGroupV3(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_dbParameterGroupV3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, t, resourceName, &dbCluster),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "timestream-influxdb", regexache.MustCompile(`db-cluster/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "db_parameter_group_identifier", "InfluxDBV3Core"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8181"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_type"),
					resource.TestCheckResourceAttrSet(resourceName, "influx_auth_parameters_secret_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrUsername, names.AttrPassword, "organization", names.AttrAllocatedStorage},
			},
		},
	})
}

func TestAccTimestreamInfluxDBDBCluster_validateConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccDBClusterConfig_v2MissingAllocatedStorage(rName),
				ExpectError: regexache.MustCompile("allocated_storage is required for InfluxDB V2 clusters"),
			},
			{
				Config:      testAccDBClusterConfig_v2MissingBucket(rName),
				ExpectError: regexache.MustCompile("bucket is required for InfluxDB V2 clusters"),
			},
			{
				Config:      testAccDBClusterConfig_v2MissingOrganization(rName),
				ExpectError: regexache.MustCompile("organization is required for InfluxDB V2 clusters"),
			},
			{
				Config:      testAccDBClusterConfig_v2MissingPassword(rName),
				ExpectError: regexache.MustCompile("password is required for InfluxDB V2 clusters"),
			},
			{
				Config:      testAccDBClusterConfig_v2MissingUsername(rName),
				ExpectError: regexache.MustCompile("username is required for InfluxDB V2 clusters"),
			},
			{
				Config:      testAccDBClusterConfig_v3WithV2Fields(rName),
				ExpectError: regexache.MustCompile(`(?s)allocated_storage must not be set when using an InfluxDB V3 db parameter.*group`),
			},
			{
				Config:      testAccDBClusterConfig_v2WithV3ParameterGroup(rName),
				ExpectError: regexache.MustCompile(`(?s)allocated_storage must not be set when using an InfluxDB V3 db parameter.*group`),
			},
			{
				Config:      testAccDBClusterConfig_v3WithAllocatedStorage(rName),
				ExpectError: regexache.MustCompile(`(?s)allocated_storage must not be set when using an InfluxDB V3 db parameter.*group`),
			},
			{
				Config:      testAccDBClusterConfig_v3WithBucket(rName),
				ExpectError: regexache.MustCompile(`(?s)bucket must not be set when using an InfluxDB V3 db parameter.*group`),
			},
			{
				Config:      testAccDBClusterConfig_v3WithDeploymentType(rName),
				ExpectError: regexache.MustCompile(`(?s)deployment_type must not be set when using an InfluxDB V3 db parameter.*group`),
			},
			{
				Config:      testAccDBClusterConfig_v3WithOrganization(rName),
				ExpectError: regexache.MustCompile(`(?s)organization must not be set when using an InfluxDB V3 db parameter.*group`),
			},
			{
				Config:      testAccDBClusterConfig_v3WithPassword(rName),
				ExpectError: regexache.MustCompile(`(?s)password must not be set when using an InfluxDB V3 db parameter.*group`),
			},
			{
				Config:      testAccDBClusterConfig_v3WithUsername(rName),
				ExpectError: regexache.MustCompile(`(?s)username must not be set when using an InfluxDB V3 db parameter.*group`),
			},
		},
	})
}

func testAccCheckDBClusterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TimestreamInfluxDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreaminfluxdb_db_cluster" {
				continue
			}

			_, err := tftimestreaminfluxdb.FindDBClusterByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingDestroyed, tftimestreaminfluxdb.ResNameDBCluster, rs.Primary.ID, err)
			}

			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingDestroyed, tftimestreaminfluxdb.ResNameDBCluster, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDBClusterExists(ctx context.Context, t *testing.T, name string, dbCluster *timestreaminfluxdb.GetDbClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBCluster, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBCluster, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).TimestreamInfluxDBClient(ctx)
		resp, err := tftimestreaminfluxdb.FindDBClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBCluster, rs.Primary.ID, err)
		}

		*dbCluster = *resp

		return nil
	}
}

func testAccPreCheckDBClusters(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).TimestreamInfluxDBClient(ctx)

	input := &timestreaminfluxdb.ListDbClustersInput{}
	_, err := conn.ListDbClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDBClusterConfig_base(rName string, subnetCount int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, subnetCount), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccDBClusterConfig_v3Base(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count = 2

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  route_table_id  = aws_route_table.test.id
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
}

resource "aws_security_group_rule" "test" {
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  prefix_list_ids   = [aws_vpc_endpoint.s3.prefix_list_id]
  security_group_id = aws_security_group.test.id
}
`, rName)
}

// Minimal configuration.
func testAccDBClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
# InfluxDB V2.
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  allocated_storage      = 20
  bucket                 = "initial"
  db_instance_type       = "db.influx.medium"
  name                   = %[1]q
  organization           = "organization"
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccDBClusterConfig_dbInstanceType(rName string, instanceType string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = %[2]q
  port                   = 8086
  bucket                 = "initial"
  organization           = "organization"
}
`, rName, instanceType))
}

// Configuration with log_delivery_configuration set and enabled.
func testAccDBClusterConfig_logDeliveryConfigurationEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["s3:PutObject"]
    principals {
      type        = "Service"
      identifiers = ["timestream-influxdb.amazonaws.com"]
    }
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  publicly_accessible    = false
  port                   = 8086
  bucket                 = "initial"
  organization           = "organization"

  log_delivery_configuration {
    s3_configuration {
      bucket_name = aws_s3_bucket.test.bucket
      enabled     = %[2]t
    }
  }
}
`, rName, enabled))
}

func testAccDBClusterConfig_publiclyAccessible(rName string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route" "test" {
  route_table_id         = aws_vpc.test.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test[0].id
  route_table_id = aws_vpc.test.main_route_table_id
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id            = aws_security_group.test.id
  referenced_security_group_id = aws_security_group.test.id
  ip_protocol                  = -1
}

resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

  publicly_accessible = true
}
`, rName))
}

func testAccDBClusterConfig_deploymentType(rName string, deploymentType string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  db_storage_type        = "InfluxIOIncludedT1"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

  deployment_type = %[2]q
}
`, rName, deploymentType))
}

func testAccDBClusterConfig_networkTypeIPV4(rName string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  port                   = 8086
  bucket                 = "initial"
  organization           = "organization"

  network_type = "IPV4"
}
`, rName))
}

func testAccDBClusterConfig_port(rName string, port string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

  port = %[2]s
}
`, rName, port))
}

func testAccDBClusterConfig_allocatedStorage(rName string, storageAmount string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

  allocated_storage = %[2]s
}
`, rName, storageAmount))
}

func testAccDBClusterConfig_dbStorageType(rName string, dbStorageType string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 400
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

  db_storage_type = %[2]q
}
`, rName, dbStorageType))
}

func testAccDBClusterConfig_failoverMode(rName string, failoverMode string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 400
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"

  failover_mode = %[2]q
}
`, rName, failoverMode))
}

func testAccDBClusterConfig_dbParameterGroupV3(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}

func testAccDBClusterConfig_v2MissingAllocatedStorage(rName string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"
  deployment_type        = "MULTI_NODE_READ_REPLICAS"
}
`, rName))
}

func testAccDBClusterConfig_v2MissingBucket(rName string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  organization           = "organization"
  deployment_type        = "MULTI_NODE_READ_REPLICAS"
}
`, rName))
}

func testAccDBClusterConfig_v2MissingOrganization(rName string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  deployment_type        = "MULTI_NODE_READ_REPLICAS"
}
`, rName))
}

func testAccDBClusterConfig_v2MissingPassword(rName string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  username               = "admin"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"
  deployment_type        = "MULTI_NODE_READ_REPLICAS"
}
`, rName))
}

func testAccDBClusterConfig_v2MissingUsername(rName string) string {
	return acctest.ConfigCompose(testAccDBClusterConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                   = %[1]q
  allocated_storage      = 20
  password               = "testpassword"
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test.id]
  db_instance_type       = "db.influx.medium"
  bucket                 = "initial"
  organization           = "organization"
  deployment_type        = "MULTI_NODE_READ_REPLICAS"
}
`, rName))
}

func testAccDBClusterConfig_v2WithV3ParameterGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  allocated_storage             = 20
  username                      = "admin"
  password                      = "testpassword"
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"
  bucket                        = "initial"
  organization                  = "organization"
  deployment_type               = "MULTI_NODE_READ_REPLICAS"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}

func testAccDBClusterConfig_v3WithV2Fields(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  allocated_storage             = 20
  username                      = "admin"
  password                      = "testpassword"
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"
  bucket                        = "initial"
  organization                  = "organization"
  deployment_type               = "MULTI_NODE_READ_REPLICAS"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}

func testAccDBClusterConfig_v3WithAllocatedStorage(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  allocated_storage             = 20
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}

func testAccDBClusterConfig_v3WithBucket(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  bucket                        = "initial"
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}

func testAccDBClusterConfig_v3WithDeploymentType(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  deployment_type               = "MULTI_NODE_READ_REPLICAS"
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}

func testAccDBClusterConfig_v3WithOrganization(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  organization                  = "organization"
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}

func testAccDBClusterConfig_v3WithPassword(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  password                      = "testpassword"
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}

func testAccDBClusterConfig_v3WithUsername(rName string) string {
	return acctest.ConfigCompose(
		testAccDBClusterConfig_base(rName, 2),
		testAccDBClusterConfig_v3Base(rName),
		fmt.Sprintf(`
resource "aws_timestreaminfluxdb_db_cluster" "test" {
  name                          = %[1]q
  username                      = "admin"
  vpc_subnet_ids                = aws_subnet.test[*].id
  vpc_security_group_ids        = [aws_security_group.test.id]
  db_instance_type              = "db.influx.medium"
  db_parameter_group_identifier = "InfluxDBV3Core"

  depends_on = [
    aws_vpc_endpoint_route_table_association.test,
    aws_security_group_rule.test,
  ]
}
`, rName))
}
