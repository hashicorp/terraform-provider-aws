package rds_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRDSEngineVersionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_rds_engine_version.test"
	engine := "oracle-ee"
	version := "19.0.0.0.ru-2020-07.rur-2020-07.r1"
	paramGroup := "oracle-ee-19"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionBasicDataSourceConfig(engine, version, paramGroup),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "version", version),
					resource.TestCheckResourceAttr(dataSourceName, "parameter_group_family", paramGroup),

					resource.TestCheckResourceAttrSet(dataSourceName, "default_character_set"),
					resource.TestCheckResourceAttrSet(dataSourceName, "engine_description"),
					resource.TestMatchResourceAttr(dataSourceName, "exportable_log_types.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestMatchResourceAttr(dataSourceName, "supported_character_sets.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_feature_names.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_modes.#", regexp.MustCompile(`^[0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_timezones.#", regexp.MustCompile(`^[0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_global_databases"),
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
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionUpgradeTargetsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "valid_upgrade_targets.#", regexp.MustCompile(`^[1-9][0-9]*`)),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionPreferredDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", "8.0.27"),
				),
			},
		},
	})
}

func TestAccRDSEngineVersionDataSource_defaultOnly(t *testing.T) {
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDefaultOnlyDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "version"),
				),
			},
		},
	})
}

func testAccEngineVersionPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	input := &rds.DescribeDBEngineVersionsInput{
		Engine:      aws.String("mysql"),
		DefaultOnly: aws.Bool(true),
	}

	_, err := conn.DescribeDBEngineVersions(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEngineVersionBasicDataSourceConfig(engine, version, paramGroup string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                 = %[1]q
  version                = %[2]q
  parameter_group_family = %[3]q
}
`, engine, version, paramGroup)
}

func testAccEngineVersionUpgradeTargetsDataSourceConfig() string {
	return `
data "aws_rds_engine_version" "test" {
  engine  = "mysql"
  version = "8.0.27"
}
`
}

func testAccEngineVersionPreferredDataSourceConfig() string {
	return `
data "aws_rds_engine_version" "test" {
  engine             = "mysql"
  preferred_versions = ["85.9.12", "8.0.27", "8.0.26"]
}
`
}

func testAccEngineVersionDefaultOnlyDataSourceConfig() string {
	return `
data "aws_rds_engine_version" "test" {
  engine = "mysql"
}
`
}
