// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSEngineVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"
	engine := tfrds.InstanceEngineOracleEnterprise
	version := "19.0.0.0.ru-2020-07.rur-2020-07.r1"
	paramGroup := "oracle-ee-19"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_basic(engine, version, paramGroup),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, engine),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, version),
					resource.TestCheckResourceAttr(dataSourceName, "version_actual", version),
					resource.TestCheckResourceAttr(dataSourceName, "parameter_group_family", paramGroup),
					resource.TestCheckResourceAttrSet(dataSourceName, "default_character_set"),
					resource.TestCheckResourceAttrSet(dataSourceName, "engine_description"),
					resource.TestMatchResourceAttr(dataSourceName, "exportable_log_types.#", regexache.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
					resource.TestMatchResourceAttr(dataSourceName, "supported_character_sets.#", regexache.MustCompile(`^[1-9][0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_feature_names.#", regexache.MustCompile(`^[1-9][0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_modes.#", regexache.MustCompile(`^[0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_timezones.#", regexache.MustCompile(`^[0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_global_databases"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_limitless_database"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_log_exports_to_cloudwatch"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_parallel_query"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_read_replica"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_description"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_upgradeTargets(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_upgradeTargets(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "valid_upgrade_targets.#", regexache.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_actual"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_preferred(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferred(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "8.0.32"),
					resource.TestCheckResourceAttr(dataSourceName, "version_actual", "8.0.32"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_preferred2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "8.0.32"),
					resource.TestCheckResourceAttr(dataSourceName, "version_actual", "8.0.32"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_preferredVersionsPreferredUpgradeTargets(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferredVersionsPreferredUpgrades(tfrds.InstanceEngineMySQL, `"5.7.37", "5.7.38", "5.7.39"`, `"8.0.34"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "5.7.39"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_preferredVersionsPreferredUpgrades(tfrds.InstanceEngineMySQL, `"5.7.44", "5.7.38", "5.7.39"`, `"8.0.32","8.0.33"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "5.7.44"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_preferredUpgradeTargetsVersion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferredUpgradeTargetsVersion(tfrds.InstanceEngineMySQL, "5.7", `"8.0.44", "8.0.35", "8.0.34"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^5\.7`)),
					resource.TestMatchResourceAttr(dataSourceName, "version_actual", regexache.MustCompile(`^5\.7\.`)),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_preferredMajorTargets(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferredMajorTarget(tfrds.InstanceEngineMySQL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^5\.7\.`)),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_preferredMajorTarget(tfrds.InstanceEngineAuroraPostgreSQL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^15\.`)),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_defaultOnlyImplicit(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_defaultOnlyImplicit(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrVersion),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_defaultOnlyExplicit(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_defaultOnlyExplicit(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^8\.0`)),
					resource.TestMatchResourceAttr(dataSourceName, "version_actual", regexache.MustCompile(`^8\.0\.`)),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_includeAll(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_includeAll(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "8.0.20"),
					resource.TestCheckResourceAttr(dataSourceName, "version_actual", "8.0.20"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_filter(tfrds.ClusterEngineAuroraPostgreSQL, "serverless"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(dataSourceName, "supported_modes.0", "serverless"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_filter(tfrds.ClusterEngineAuroraPostgreSQL, "global"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(dataSourceName, "supported_modes.0", "global"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_latest(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_latest(true, `"13.9", "12.7", "11.12", "15.4", "10.17", "9.6.22"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "15.4"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_latest(false, `"13.9", "12.7", "11.12", "15.4", "10.17", "9.6.22"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "13.9"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_latest2(tfrds.InstanceEngineAuroraPostgreSQL, "15"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^15`)),
					resource.TestMatchResourceAttr(dataSourceName, "version_actual", regexache.MustCompile(`^15\.[0-9]`)),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_latest2(tfrds.InstanceEngineMySQL, "8.0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^8\.0`)),
					resource.TestMatchResourceAttr(dataSourceName, "version_actual", regexache.MustCompile(`^8\.0\.[0-9]+$`)),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_hasMinorMajor(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_hasMajorMinorTarget(tfrds.InstanceEngineAuroraPostgreSQL, true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(dataSourceName, "valid_major_targets.#", func(value string) error {
						intValue, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("could not convert string to int: %v", err)
						}

						if intValue <= 0 {
							return fmt.Errorf("value is not greater than 0: %d", intValue)
						}

						return nil
					}),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_hasMajorMinorTarget(tfrds.InstanceEngineAuroraPostgreSQL, false, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(dataSourceName, "valid_minor_targets.#", func(value string) error {
						intValue, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("could not convert string to int: %v", err)
						}

						if intValue <= 0 {
							return fmt.Errorf("value is not greater than 0: %d", intValue)
						}

						return nil
					}),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_hasMajorMinorTarget(tfrds.InstanceEngineAuroraPostgreSQL, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(dataSourceName, "valid_major_targets.#", func(value string) error {
						intValue, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("could not convert string to int: %v", err)
						}

						if intValue <= 0 {
							return fmt.Errorf("value is not greater than 0: %d", intValue)
						}

						return nil
					}),
					resource.TestCheckResourceAttrWith(dataSourceName, "valid_minor_targets.#", func(value string) error {
						intValue, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("could not convert string to int: %v", err)
						}

						if intValue <= 0 {
							return fmt.Errorf("value is not greater than 0: %d", intValue)
						}

						return nil
					}),
				),
			},
		},
	})
}

func testAccEngineVersionPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeDBEngineVersionsInput{
		Engine:      aws.String(tfrds.InstanceEngineMySQL),
		DefaultOnly: aws.Bool(true),
	}

	_, err := conn.DescribeDBEngineVersionsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEngineVersionDataSourceConfig_basic(engine, version, paramGroup string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                 = %[1]q
  version                = %[2]q
  parameter_group_family = %[3]q
}
`, engine, version, paramGroup)
}

func testAccEngineVersionDataSourceConfig_upgradeTargets() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine  = %[1]q
  version = "8.0.32"
}
`, tfrds.InstanceEngineMySQL)
}

func testAccEngineVersionDataSourceConfig_preferred() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = %[1]q
  preferred_versions = ["85.9.12", "8.0.32", "8.0.31"]
}
`, tfrds.InstanceEngineMySQL)
}

func testAccEngineVersionDataSourceConfig_preferred2() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = %[1]q
  version            = "8.0"
  preferred_versions = ["85.9.12", "8.0.32", "8.0.31"]
}
`, tfrds.InstanceEngineMySQL)
}

func testAccEngineVersionDataSourceConfig_preferredVersionsPreferredUpgrades(engine, preferredVersions, preferredUpgrades string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                    = %[1]q
  latest                    = true
  preferred_versions        = [%[2]s]
  preferred_upgrade_targets = [%[3]s]
}
`, engine, preferredVersions, preferredUpgrades)
}

func testAccEngineVersionDataSourceConfig_preferredUpgradeTargetsVersion(engine, version, preferredUpgrades string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                    = %[1]q
  version                   = %[2]q
  preferred_upgrade_targets = [%[3]s]
}
`, engine, version, preferredUpgrades)
}

func testAccEngineVersionDataSourceConfig_preferredMajorTarget(engine string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "latest" {
  engine = %[1]q
  latest = true
}

data "aws_rds_engine_version" "test" {
  engine                  = %[1]q
  latest                  = true
  preferred_major_targets = [data.aws_rds_engine_version.latest.version]
}
`, engine)
}

func testAccEngineVersionDataSourceConfig_defaultOnlyImplicit() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine = %[1]q
}
`, tfrds.InstanceEngineMySQL)
}

func testAccEngineVersionDataSourceConfig_defaultOnlyExplicit() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine       = %[1]q
  version      = "8.0"
  default_only = true
}
`, tfrds.InstanceEngineMySQL)
}

func testAccEngineVersionDataSourceConfig_includeAll() string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine      = %[1]q
  version     = "8.0.20"
  include_all = true
}
`, tfrds.InstanceEngineMySQL)
}

func testAccEngineVersionDataSourceConfig_filter(engine, engineMode string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine      = %[1]q
  latest      = true
  include_all = true

  filter {
    name   = "engine-mode"
    values = [%[2]q]
  }
}
`, engine, engineMode)
}

func testAccEngineVersionDataSourceConfig_latest(latest bool, preferredVersions string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = %[1]q
  latest             = %[2]t
  preferred_versions = [%[3]s]
}
`, tfrds.InstanceEngineAuroraPostgreSQL, latest, preferredVersions)
}

func testAccEngineVersionDataSourceConfig_latest2(engine, majorVersion string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine  = %[1]q
  version = %[2]q
  latest  = true
}
`, engine, majorVersion)
}

func testAccEngineVersionDataSourceConfig_hasMajorMinorTarget(engine string, major, minor bool) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine           = %[1]q
  has_major_target = %[2]t
  has_minor_target = %[3]t
  latest           = true
}
`, engine, major, minor)
}
