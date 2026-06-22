// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneEngineVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"
	dataSourceNameLatest := "data.aws_neptune_engine_version.latest"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_basic(tfneptune.DefaultEngine),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, tfneptune.DefaultEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, dataSourceNameLatest, names.AttrVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, "version_actual", dataSourceNameLatest, "version_actual"),
					resource.TestCheckResourceAttrSet(dataSourceName, "parameter_group_family"),
					resource.TestCheckResourceAttrSet(dataSourceName, "engine_description"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_global_databases"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_log_exports_to_cloudwatch"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_read_replica"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_description"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("exportable_log_types"), tfknownvalue.ListNotEmpty()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("supported_timezones"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_upgradeTargets(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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

func TestAccNeptuneEngineVersionDataSource_preferred(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"
	dataSourceNameLatest := "data.aws_neptune_engine_version.latest"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferred(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, dataSourceNameLatest, "version_actual"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version_actual", dataSourceNameLatest, "version_actual"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_preferred2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, dataSourceNameLatest, "version_actual"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version_actual", dataSourceNameLatest, "version_actual"),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_preferredVersionsPreferredUpgradeTargets(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferredVersionsPreferredUpgrades(tfneptune.DefaultEngine, `"1.4.1.0", "1.4.2.0", "1.4.3.0"`, `"1.4.5.0"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "1.4.3.0"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_preferredVersionsPreferredUpgrades(tfneptune.DefaultEngine, `"1.3.2.0", "1.3.3.0", "1.4.0.0"`, `"1.4.4.0","1.4.5.0"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "1.4.0.0"),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_preferredUpgradeTargetsVersion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferredUpgradeTargetsVersion(tfneptune.DefaultEngine, "1.3", `"1.4.0.0", "1.4.1.0", "1.4.2.0"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`^1\.3`)),
					resource.TestMatchResourceAttr(dataSourceName, "version_actual", regexache.MustCompile(`^1\.3\.`)),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_preferredMajorTargets(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"

	majorTarget := "1.4"
	oneLess := `^1\.3\.`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferredMajorTarget(tfneptune.DefaultEngine, majorTarget),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(oneLess)),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_defaultOnlyImplicit(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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

func TestAccNeptuneEngineVersionDataSource_defaultOnlyExplicit(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"
	updateDataSourceName := "data.aws_neptune_engine_version.latest"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_defaultOnlyExplicit(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "version_actual", updateDataSourceName, "version_actual"),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_latest(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"
	dataSourceNameLatest := "data.aws_neptune_engine_version.latest"
	dataSourceNameEarlier2 := "data.aws_neptune_engine_version.earlier2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_latest(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "version_actual", dataSourceNameLatest, "version_actual"),
				),
			},
			{
				Config: testAccEngineVersionDataSourceConfig_latest(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "version_actual", dataSourceNameEarlier2, "version_actual"),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_hasMinorMajor(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_neptune_engine_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_hasMajorMinorTarget(tfneptune.DefaultEngine, true, false),
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
				Config: testAccEngineVersionDataSourceConfig_hasMajorMinorTarget(tfneptune.DefaultEngine, false, true),
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
				Config: testAccEngineVersionDataSourceConfig_hasMajorMinorTarget(tfneptune.DefaultEngine, true, true),
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
	conn := acctest.ProviderMeta(ctx, t).NeptuneClient(ctx)

	input := &neptune.DescribeDBEngineVersionsInput{
		Engine:      aws.String(tfneptune.DefaultEngine),
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

func testAccEngineVersionDataSourceConfig_basic(engine string) string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine  = %[1]q
  version = data.aws_neptune_engine_version.latest.version_actual
}

data "aws_neptune_engine_version" "latest" {
  engine = %[1]q
  latest = true
}
`, engine)
}

func testAccEngineVersionDataSourceConfig_upgradeTargets() string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine                    = %[1]q
  latest                    = true
  preferred_upgrade_targets = [data.aws_neptune_engine_version.latest.version_actual]
}

data "aws_neptune_engine_version" "latest" {
  engine = %[1]q
  latest = true
}
`, tfneptune.DefaultEngine)
}

func testAccEngineVersionDataSourceConfig_preferred() string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine             = %[1]q
  preferred_versions = ["85.9.12", data.aws_neptune_engine_version.latest.version_actual]
}

data "aws_neptune_engine_version" "latest" {
  engine = %[1]q
  latest = true
}
`, tfneptune.DefaultEngine)
}

func testAccEngineVersionDataSourceConfig_preferred2() string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine             = %[1]q
  version            = join(".", slice(split(".", data.aws_neptune_engine_version.latest.version_actual), 0, 2))
  preferred_versions = ["85.9.12", data.aws_neptune_engine_version.latest.version_actual]
}

data "aws_neptune_engine_version" "latest" {
  engine = %[1]q
  latest = true
}
`, tfneptune.DefaultEngine)
}

func testAccEngineVersionDataSourceConfig_preferredVersionsPreferredUpgrades(engine, preferredVersions, preferredUpgrades string) string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine                    = %[1]q
  latest                    = true
  preferred_versions        = [%[2]s]
  preferred_upgrade_targets = [%[3]s]
}
`, engine, preferredVersions, preferredUpgrades)
}

func testAccEngineVersionDataSourceConfig_preferredUpgradeTargetsVersion(engine, version, preferredUpgrades string) string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine                    = %[1]q
  version                   = %[2]q
  preferred_upgrade_targets = [%[3]s]
}
`, engine, version, preferredUpgrades)
}

func testAccEngineVersionDataSourceConfig_preferredMajorTarget(engine, majorVersion string) string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "latest" {
  engine  = %[1]q
  version = %[2]q
  latest  = true
}

data "aws_neptune_engine_version" "test" {
  engine                  = %[1]q
  latest                  = true
  preferred_major_targets = [data.aws_neptune_engine_version.latest.version]
}
`, engine, majorVersion)
}

func testAccEngineVersionDataSourceConfig_defaultOnlyImplicit() string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine = %[1]q
}
`, tfneptune.DefaultEngine)
}

func testAccEngineVersionDataSourceConfig_defaultOnlyExplicit() string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine       = %[1]q
  version      = join(".", slice(split(".", data.aws_neptune_engine_version.latest.version_actual), 0, 2))
  default_only = true
}

data "aws_neptune_engine_version" "latest" {
  engine = %[1]q
  latest = true
}
`, tfneptune.DefaultEngine)
}

func testAccEngineVersionDataSourceConfig_latest(latest bool) string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine = %[1]q
  latest = %[2]t

  preferred_versions = [
    data.aws_neptune_engine_version.earlier2.version_actual,
    data.aws_neptune_engine_version.earlier1.version_actual,
    data.aws_neptune_engine_version.latest.version_actual
  ]
}

data "aws_neptune_engine_version" "earlier2" {
  engine                    = %[1]q
  latest                    = true
  preferred_upgrade_targets = [data.aws_neptune_engine_version.earlier1.version_actual]
}

data "aws_neptune_engine_version" "earlier1" {
  engine                    = %[1]q
  latest                    = true
  preferred_upgrade_targets = [data.aws_neptune_engine_version.latest.version_actual]
}

data "aws_neptune_engine_version" "latest" {
  engine = %[1]q
  latest = true
}
`, tfneptune.DefaultEngine, latest)
}

func testAccEngineVersionDataSourceConfig_hasMajorMinorTarget(engine string, major, minor bool) string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine           = %[1]q
  has_major_target = %[2]t
  has_minor_target = %[3]t
  latest           = true
}
`, engine, major, minor)
}
