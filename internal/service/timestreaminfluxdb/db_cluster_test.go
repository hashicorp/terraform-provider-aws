// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftimestreaminfluxdb "github.com/hashicorp/terraform-provider-aws/internal/service/timestreaminfluxdb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamInfluxDBDBCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "timestream-influxdb", regexache.MustCompile(`db-cluster/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "db_storage_type", string(awstypes.DbStorageTypeInfluxIoIncludedT1)),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.ClusterDeploymentTypeMultiNodeReadReplicas)),
					resource.TestCheckResourceAttr(resourceName, "failover_mode", string(awstypes.FailoverModeAutomatic)),
					resource.TestCheckResourceAttrSet(resourceName, "influx_auth_parameters_secret_arn"),
					resource.TestCheckResourceAttr(resourceName, "network_type", string(awstypes.NetworkTypeIpv4)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8086"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
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

func TestAccTimestreamInfluxDBDBCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftimestreaminfluxdb.ResourceDBCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_dbInstanceType(rName, string(awstypes.DbInstanceTypeDbInfluxMedium)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster1),
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
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckDBClusterNotRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "db_instance_type", string(awstypes.DbInstanceTypeDbInfluxLarge)),
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

func TestAccTimestreamInfluxDBDBCluster_logDeliveryConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster1, dbCluster2 timestreaminfluxdb.GetDbClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_logDeliveryConfigurationEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster1),
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
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckDBClusterNotRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.s3_configuration.0.enabled", acctest.CtFalse),
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

func TestAccTimestreamInfluxDBDBCluster_networkType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_networkTypeIPV4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_port(rName, port1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster1),
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
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckDBClusterNotRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, port2),
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

func TestAccTimestreamInfluxDBDBCluster_allocatedStorage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbCluster timestreaminfluxdb.GetDbClusterOutput
	allocatedStorage := "20"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_allocatedStorage(rName, allocatedStorage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_dbStorageType(rName, string(awstypes.DbStorageTypeInfluxIoIncludedT1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_publiclyAccessible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_deploymentType(rName, string(awstypes.ClusterDeploymentTypeMultiNodeReadReplicas)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreaminfluxdb_db_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBClusters(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDBClusterConfig_failoverMode(rName, string(awstypes.FailoverModeAutomatic)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster1),
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
					testAccCheckDBClusterExists(ctx, resourceName, &dbCluster2),
					testAccCheckDBClusterNotRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "failover_mode", string(awstypes.FailoverModeNoFailover)),
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

func testAccCheckDBClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamInfluxDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreaminfluxdb_db_cluster" {
				continue
			}

			_, err := tftimestreaminfluxdb.FindDBClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckDBClusterExists(ctx context.Context, name string, dbCluster *timestreaminfluxdb.GetDbClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBCluster, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBCluster, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamInfluxDBClient(ctx)
		resp, err := tftimestreaminfluxdb.FindDBClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingExistence, tftimestreaminfluxdb.ResNameDBCluster, rs.Primary.ID, err)
		}

		*dbCluster = *resp

		return nil
	}
}

func testAccPreCheckDBClusters(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamInfluxDBClient(ctx)

	input := &timestreaminfluxdb.ListDbClustersInput{}
	_, err := conn.ListDbClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckDBClusterNotRecreated(before, after *timestreaminfluxdb.GetDbClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.TimestreamInfluxDB, create.ErrActionCheckingNotRecreated, 
                tftimestreaminfluxdb.ResNameDBCluster, 
                fmt.Sprintf("before: %s, after: %s", beforeID, afterID), 
                errors.New("resource was recreated when it should have been updated in-place"))
		}

		return nil
	}
}

func testAccDBClusterConfig_base(rName string, subnetCount int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, subnetCount), `
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}
`)
}

// Minimal configuration.
func testAccDBClusterConfig_basic(rName string) string {
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
