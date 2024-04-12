// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSReplicationConfig_basic(t *testing.T) {
	t.Parallel()

	for _, migrationType := range dms.MigrationTypeValue_Values() {
		t.Run(migrationType, func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
			resourceName := "aws_dms_replication_config.test"

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccReplicationConfigConfig_basic(rName, migrationType),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckReplicationConfigExists(ctx, resourceName),
							resource.TestCheckResourceAttrSet(resourceName, "arn"),
							acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dms", regexache.MustCompile(`replication-config:[A-Z0-9]{26}`)),
							resource.TestCheckResourceAttr(resourceName, "compute_config.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.availability_zone", ""),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.dns_name_servers", ""),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.kms_key_id", ""),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "128"),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "2"),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.multi_az", "false"),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.preferred_maintenance_window", "sun:23:45-mon:00:30"),
							resource.TestCheckResourceAttrSet(resourceName, "compute_config.0.replication_subnet_group_id"),
							resource.TestCheckResourceAttr(resourceName, "compute_config.0.vpc_security_group_ids.#", "0"),
							resource.TestCheckResourceAttr(resourceName, "replication_config_identifier", rName),
							acctest.CheckResourceAttrEquivalentJSON(resourceName, "replication_settings", defaultReplicationConfigSettings[migrationType]),
							resource.TestCheckResourceAttr(resourceName, "replication_type", migrationType),
							resource.TestCheckNoResourceAttr(resourceName, "resource_identifier"),
							resource.TestCheckResourceAttrPair(resourceName, "source_endpoint_arn", "aws_dms_endpoint.source", "endpoint_arn"),
							resource.TestCheckResourceAttr(resourceName, "start_replication", "false"),
							resource.TestCheckResourceAttr(resourceName, "supplemental_settings", ""),
							acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", "1"),
							resource.TestCheckResourceAttrPair(resourceName, "target_endpoint_arn", "aws_dms_endpoint.target", "endpoint_arn"),
							resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
							resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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

// func TestAccDMSReplicationConfig_noChangeOnDefault(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_dms_replication_config.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
// 		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccReplicationConfigConfig_noChangeOnDefault(rName),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					testAccCheckReplicationConfigExists(ctx, resourceName),
// 					resource.TestCheckResourceAttrSet(resourceName, "arn"),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.#", "1"),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.0.availability_zone", ""),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.0.dns_name_servers", ""),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.0.kms_key_id", ""),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "128"),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "2"),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.0.multi_az", "false"),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.0.preferred_maintenance_window", "sun:23:45-mon:00:30"),
// 					resource.TestCheckResourceAttrSet(resourceName, "compute_config.0.replication_subnet_group_id"),
// 					resource.TestCheckResourceAttr(resourceName, "compute_config.0.vpc_security_group_ids.#", "0"),
// 					resource.TestCheckResourceAttr(resourceName, "replication_config_identifier", rName),
// 					resource.TestCheckResourceAttrSet(resourceName, "replication_settings"),
// 					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
// 					resource.TestCheckNoResourceAttr(resourceName, "resource_identifier"),
// 					resource.TestCheckResourceAttrSet(resourceName, "source_endpoint_arn"),
// 					resource.TestCheckResourceAttr(resourceName, "start_replication", "false"),
// 					resource.TestCheckResourceAttr(resourceName, "supplemental_settings", ""),
// 					resource.TestCheckResourceAttrSet(resourceName, "table_mappings"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
// 					resource.TestCheckResourceAttrSet(resourceName, "target_endpoint_arn"),
// 				),
// 			},
// 			{
// 				ResourceName:            resourceName,
// 				ImportState:             true,
// 				ImportStateVerify:       true,
// 				ImportStateVerifyIgnore: []string{"start_replication", "resource_identifier"},
// 			},
// 		},
// 	})
// }

func TestAccDMSReplicationConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_basic(rName, "cdc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdms.ResourceReplicationConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDMSReplicationConfig_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccReplicationConfigConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccReplicationConfigConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDMSReplicationConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_update(rName, "cdc", 2, 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "2"),
				),
			},
			{
				Config: testAccReplicationConfigConfig_update(rName, "cdc", 4, 32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "32"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "4"),
				),
			},
		},
	})
}

func TestAccDMSReplicationConfig_startReplication(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_startReplication(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "start_replication", "true"),
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
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "start_replication", "false"),
				),
			},
		},
	})
}

func testAccCheckReplicationConfigExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn(ctx)

		_, err := tfdms.FindReplicationConfigByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckReplicationConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_replication_config" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn(ctx)

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

func testAccReplicationConfigConfig_noChangeOnDefault(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationConfigConfig_base_DummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  replication_settings          = "{\"Logging\":{\"EnableLogging\":true}}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName))
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
