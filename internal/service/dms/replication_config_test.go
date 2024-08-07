// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSReplicationConfig_basic(t *testing.T) {
	t.Parallel()

	for _, migrationType := range enum.Values[awstypes.MigrationTypeValue]() { //nolint:paralleltest // false positive
		t.Run(migrationType, func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
			resourceName := "aws_dms_replication_config.test"
			var v awstypes.ReplicationConfig

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccReplicationConfigConfig_basic(rName, migrationType),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckReplicationConfigExists(ctx, resourceName, &v),
							resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
							acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "dms", regexache.MustCompile(`replication-config:[A-Z0-9]{26}`)),
							resource.TestCheckResourceAttr(resourceName, "compute_config.#", acctest.Ct1),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.availability_zone", ""),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.dns_name_servers", ""),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.kms_key_id", ""),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "128"),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.multi_az", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.preferred_maintenance_window", "sun:23:45-mon:00:30"),
							resource.TestCheckResourceAttrSet(resourceName, "compute_config.0.replication_subnet_group_id"),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.vpc_security_group_ids.#", acctest.Ct0),
							resource.TestCheckResourceAttr(resourceName, "replication_config_identifier", rName),
							acctest.CheckResourceAttrEquivalentJSON(resourceName, "replication_settings", defaultReplicationConfigSettings[migrationType]),
							resource.TestCheckResourceAttr(resourceName, "replication_type", migrationType),
							resource.TestCheckNoResourceAttr(resourceName, "resource_identifier"),
							resource.TestCheckResourceAttrPair(resourceName, "source_endpoint_arn", "aws_dms_endpoint.source", "endpoint_arn"),
							resource.TestCheckResourceAttr(resourceName, "start_replication", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "supplemental_settings", ""),
							acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", acctest.Ct1),
							resource.TestCheckResourceAttrPair(resourceName, "target_endpoint_arn", "aws_dms_endpoint.target", "endpoint_arn"),
							resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
							resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
						),
					},
					{
						ResourceName:            resourceName,
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"start_replication", "resource_identifier"},
					},
				},
			})
		})
	}
}

func TestAccDMSReplicationConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"
	var v awstypes.ReplicationConfig

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_basic(rName, "cdc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdms.ResourceReplicationConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDMSReplicationConfig_settings_EnableLogging(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"
	var v awstypes.ReplicationConfig

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_settings_EnableLogging(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.EnableLogging", acctest.CtTrue),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.EnableLogContext", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.LogComponents[?Id=='DATA_STRUCTURE'].Severity | [0]", "LOGGER_SEVERITY_DEFAULT"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "type(Logging.CloudWatchLogGroup)", "null"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "type(Logging.CloudWatchLogStream)", "null"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication"},
			},
			{
				Config: testAccReplicationConfigConfig_settings_EnableLogContext(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.EnableLogging", acctest.CtTrue),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.EnableLogContext", acctest.CtTrue),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.LogComponents[?Id=='DATA_STRUCTURE'].Severity | [0]", "LOGGER_SEVERITY_DEFAULT"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "type(Logging.CloudWatchLogGroup)", "null"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "type(Logging.CloudWatchLogStream)", "null"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication"},
			},
			{
				Config: testAccReplicationConfigConfig_settings_EnableLogging(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.EnableLogging", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.EnableLogContext", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.LogComponents[?Id=='DATA_STRUCTURE'].Severity | [0]", "LOGGER_SEVERITY_DEFAULT"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "type(Logging.CloudWatchLogGroup)", "null"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "type(Logging.CloudWatchLogStream)", "null"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication"},
			},
		},
	})
}

func TestAccDMSReplicationConfig_settings_LoggingValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationConfigConfig_settings_EnableLogContext(rName, false, true),
				ExpectError: regexache.MustCompile(`The parameter Logging.EnableLogContext is not allowed when\s+Logging.EnableLogging is not set to true.`),
			},
			{
				Config:      testAccReplicationConfigConfig_settings_LoggingReadOnly(rName, "CloudWatchLogGroup"),
				ExpectError: regexache.MustCompile(`The parameter Logging.CloudWatchLogGroup is read-only and cannot be set.`),
			},
			{
				Config:      testAccReplicationConfigConfig_settings_LoggingReadOnly(rName, "CloudWatchLogStream"),
				ExpectError: regexache.MustCompile(`The parameter Logging.CloudWatchLogStream is read-only and cannot be set.`),
			},
		},
	})
}

func TestAccDMSReplicationConfig_settings_LogComponents(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"
	var v awstypes.ReplicationConfig

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_settings_LogComponents(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.EnableLogging", acctest.CtTrue),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.EnableLogContext", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "Logging.LogComponents[?Id=='DATA_STRUCTURE'].Severity | [0]", "LOGGER_SEVERITY_WARNING"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "type(Logging.CloudWatchLogGroup)", "null"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "type(Logging.CloudWatchLogStream)", "null"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication"},
			},
		},
	})
}

func TestAccDMSReplicationConfig_settings_StreamBuffer(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"
	var v awstypes.ReplicationConfig

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_settings_StreamBuffer(rName, 4, 16),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "StreamBufferSettings.StreamBufferCount", acctest.Ct4),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "StreamBufferSettings.StreamBufferSizeInMB", "16"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_settings", "StreamBufferSettings.CtrlStreamBufferSizeInMB", "5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication"},
			},
		},
	})
}

func TestAccDMSReplicationConfig_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"
	var v awstypes.ReplicationConfig

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccReplicationConfigConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccReplicationConfigConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDMSReplicationConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"
	var v awstypes.ReplicationConfig

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_update(rName, "cdc", 2, 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", acctest.Ct2),
				),
			},
			{
				Config: testAccReplicationConfigConfig_update(rName, "cdc", 4, 32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "32"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", acctest.Ct4),
				),
			},
		},
	})
}

func TestAccDMSReplicationConfig_startReplication(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"
	var v awstypes.ReplicationConfig

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_startReplication(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "start_replication", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication", "resource_identifier"},
			},
			{
				Config: testAccReplicationConfigConfig_startReplication(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "start_replication", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckReplicationConfigExists(ctx context.Context, n string, v *awstypes.ReplicationConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSClient(ctx)

		output, err := tfdms.FindReplicationConfigByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckReplicationConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_replication_config" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).DMSClient(ctx)

			_, err := tfdms.FindReplicationConfigByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Replication Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// testAccReplicationConfigConfig_base_DummyDatabase creates Replication Endpoints referencing valid databases.
// This should only be used in cases where actual replication is started, since it requires approcimately
// six more minutes for setup and teardown.
func testAccReplicationConfigConfig_base_ValidDatabase(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_rdsClusterBase(rName), fmt.Sprintf(`
resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test"
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_endpoint" "target" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-target"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.target.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}

resource "aws_dms_endpoint" "source" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-source"
  endpoint_type = "source"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.source.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}
`, rName))
}

// testAccReplicationConfigConfig_base_DummyDatabase creates Replication Endpoints referencing dummy databases.
// This should be used in all cases where actual replication is not started, since it shaves approcimately
// six minutes off setup and teardown time.
func testAccReplicationConfigConfig_base_DummyDatabase(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationEndpointConfig_dummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test"
  subnet_ids                           = aws_subnet.test[*].id
}
`, rName))
}

func testAccReplicationConfigConfig_basic(rName, migrationType string) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = %[2]q
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName, migrationType))
}

func testAccReplicationConfigConfig_settings_EnableLogging(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "full-load"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  # terrafmt can't handle this using jsonencode or a heredoc
  replication_settings = "{\"Logging\":{\"EnableLogging\":%[2]t}}"
}
`, rName, enabled))
}

func testAccReplicationConfigConfig_settings_EnableLogContext(rName string, enableLogging, enableLogContext bool) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "full-load"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  # terrafmt can't handle this using jsonencode or a heredoc
  replication_settings = "{\"Logging\":{\"EnableLogging\":%[2]t,\"EnableLogContext\":%[3]t}}"
}
`, rName, enableLogging, enableLogContext))
}

func testAccReplicationConfigConfig_settings_LoggingReadOnly(rName, field string) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "full-load"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  # terrafmt can't handle this using jsonencode or a heredoc
  replication_settings = "{\"Logging\":{\"EnableLogging\":true, \"%[2]s\":\"value\"}}"
}
`, rName, field))
}

func testAccReplicationConfigConfig_settings_LogComponents(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "full-load"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  replication_settings = jsonencode(
    {
      Logging = {
        EnableLogging = true,
        LogComponents = [{
          Id       = "DATA_STRUCTURE",
          Severity = "LOGGER_SEVERITY_WARNING"
        }]
      }
    }
  )
}
`, rName))
}

func testAccReplicationConfigConfig_settings_StreamBuffer(rName string, bufferCount, bufferSize int) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "full-load"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  # terrafmt can't handle this using jsonencode or a heredoc
  replication_settings = "{\"StreamBufferSettings\":{\"StreamBufferCount\":%[2]d,\"StreamBufferSizeInMB\":%[3]d}}"
}
`, rName, bufferCount, bufferSize))
}

func testAccReplicationConfigConfig_update(rName, replicationType string, minCapacity, maxCapacity int) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  resource_identifier           = %[1]q
  replication_type              = %[2]q
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "%[3]d"
    min_capacity_units           = "%[4]d"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName, replicationType, maxCapacity, minCapacity))
}

func testAccReplicationConfigConfig_startReplication(rName string, start bool) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_ValidDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  resource_identifier           = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  start_replication = %[2]t

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  depends_on = [aws_rds_cluster_instance.source, aws_rds_cluster_instance.target]
}
`, rName, start))
}

func testAccReplicationConfigConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccReplicationConfigConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

var (
	defaultReplicationConfigSettings = map[string]string{
		"cdc":               defaultReplicationConfigCdcSettings,
		"full-load":         defaultReplicationConfigFullLoadSettings,
		"full-load-and-cdc": defaultReplicationConfigFullLoadAndCdcSettings,
	}

	//go:embed testdata/replication_config/defaults/cdc.json
	defaultReplicationConfigCdcSettings string

	//go:embed testdata/replication_config/defaults/full-load.json
	defaultReplicationConfigFullLoadSettings string

	//go:embed testdata/replication_config/defaults/full-load-and-cdc.json
	defaultReplicationConfigFullLoadAndCdcSettings string
)
