// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSEngineVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"
	engine := tfrds.InstanceEngineOracleEnterprise
	version := "19.0.0.0.ru-2020-07.rur-2020-07.r1"
	paramGroup := "oracle-ee-19"

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_certificate_rotation_without_restart"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_global_databases"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_integrations"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_limitless_database"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_local_write_forwarding"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_log_exports_to_cloudwatch"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_parallel_query"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_read_replica"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_description"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("exportable_log_types"), tfknownvalue.ListNotEmpty()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("supported_character_sets"), tfknownvalue.ListNotEmpty()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("supported_feature_names"), tfknownvalue.ListNotEmpty()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("supported_modes"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("supported_timezones"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_upgradeTargets(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_upgradeTargets(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "version_actual"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("valid_upgrade_targets"), tfknownvalue.ListNotEmpty()),
				},
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_preferred(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferred_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "8.4.7"),
					resource.TestCheckResourceAttr(dataSourceName, "version_actual", "8.4.7"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_preferred_partialVersion(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "8.0.44"),
					resource.TestCheckResourceAttr(dataSourceName, "version_actual", "8.0.44"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_preferredVersionsPreferredUpgradeTargets(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferredVersionsPreferredUpgrades(tfrds.InstanceEngineMySQL, `"8.4.4", "8.4.5"`, `"8.4.7"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "8.4.5"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_preferredVersionsPreferredUpgrades(tfrds.InstanceEngineMySQL, `"8.0.42", "8.0.43", "8.0.44"`, `"8.4.6", "8.4.7"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "8.0.44"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_preferredUpgradeTargetsVersion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferredMajorTarget(tfrds.InstanceEngineMySQL),
				Check: resource.ComposeTestCheckFunc(
					// resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^8\.4\.`)),
					// As of 2026-01-15, the latest 8.0.x was *created* after the latest 8.4.x, so `latest` unfortunately picks it
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^8\.0\.`)),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_preferredMajorTarget(tfrds.InstanceEngineAuroraPostgreSQL),
				Check: resource.ComposeTestCheckFunc(
					// resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^18\.`)),
					// As of 2026-01-15, the latest 16.x was *created* after the latest 18.x, so `latest` unfortunately picks it
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^16\.`)),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_defaultOnlyImplicit(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

func TestAccRDSEngineVersionDataSource_latest_FromPreferredVersions(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_latest_FromPreferredVersions(true, `"16.10", "15.15", "14.19", "13.23", "18.1", "10.17"`),
				Check: resource.ComposeTestCheckFunc(
					// resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "18.1"),
					// Version 16.10 was *created* after 18.1, so `latest` unfortunately picks it
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "16.10"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_latest_FromPreferredVersions(false, `"16.10", "15.15", "14.19", "13.23", "18.1", "10.17"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "16.10"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_latest_OfVersion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_latest_OfVersion(tfrds.InstanceEngineAuroraPostgreSQL, "15"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^15`)),
					resource.TestMatchResourceAttr(dataSourceName, "version_actual", regexache.MustCompile(`^15\.[0-9]`)),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_latest_OfVersion(tfrds.InstanceEngineMySQL, "8.0"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
							return fmt.Errorf("could not convert string to int: %w", err)
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
							return fmt.Errorf("could not convert string to int: %w", err)
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
							return fmt.Errorf("could not convert string to int: %w", err)
						}

						if intValue <= 0 {
							return fmt.Errorf("value is not greater than 0: %d", intValue)
						}

						return nil
					}),
					resource.TestCheckResourceAttrWith(dataSourceName, "valid_minor_targets.#", func(value string) error {
						intValue, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("could not convert string to int: %w", err)
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
	conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

	input := &rds.DescribeDBEngineVersionsInput{
		Engine:      aws.String(tfrds.InstanceEngineMySQL),
		DefaultOnly: aws.Bool(true),
	}

	_, err := conn.DescribeDBEngineVersions(ctx, input)

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
  version = "8.4.6" # As of 2026-01-15, latest bug fix version with upgrade targets. End of support is 2026-09-30.
}
`, tfrds.InstanceEngineMySQL)
}

func testAccEngineVersionDataSourceConfig_preferred_basic() string {
	// Versions
	// 85.9.12: does not exist
	// 8.4.7: latest as of 2026-01-15. End of support is 2026-11-30.
	// 8.4.6: End of support is 2026-09-30.
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = %[1]q
  preferred_versions = ["85.9.12", "8.4.7", "8.4.6"]
}
`, tfrds.InstanceEngineMySQL)
}

func testAccEngineVersionDataSourceConfig_preferred_partialVersion() string {
	// Versions
	// 8.4.7: latest as of 2026-01-15. End of support is 2026-11-30.
	// 8.0.44: lastest 8.0.x as of 2026-01-15. End of support is 2026-07-31.
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = %[1]q
  version            = "8.0"
  preferred_versions = ["8.4.7", "8.0.44"]
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

func testAccEngineVersionDataSourceConfig_latest_FromPreferredVersions(latest bool, preferredVersions string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = %[1]q
  latest             = %[2]t
  preferred_versions = [%[3]s]
}
`, tfrds.InstanceEngineAuroraPostgreSQL, latest, preferredVersions)
}

func testAccEngineVersionDataSourceConfig_latest_OfVersion(engine, majorVersion string) string {
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
