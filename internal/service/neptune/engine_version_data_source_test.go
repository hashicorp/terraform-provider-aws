package neptune_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccNeptuneEngineVersionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_neptune_engine_version.test"
	version := "1.0.2.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_basic(version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", "neptune"),
					resource.TestCheckResourceAttr(dataSourceName, "version", version),
					resource.TestCheckResourceAttrSet(dataSourceName, "engine_description"),
					resource.TestMatchResourceAttr(dataSourceName, "exportable_log_types.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "parameter_group_family"),
					resource.TestMatchResourceAttr(dataSourceName, "supported_timezones.#", regexp.MustCompile(`^[0-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_log_exports_to_cloudwatch"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_read_replica"),
					resource.TestMatchResourceAttr(dataSourceName, "valid_upgrade_targets.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_description"),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_neptune_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferred(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", "neptune"),
					resource.TestCheckResourceAttr(dataSourceName, "version", "1.0.3.0"),
				),
			},
		},
	})
}

func TestAccNeptuneEngineVersionDataSource_defaultOnly(t *testing.T) {
	dataSourceName := "data.aws_neptune_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_defaultOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", "neptune"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version"),
				),
			},
		},
	})
}

func testAccEngineVersionPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn

	input := &neptune.DescribeDBEngineVersionsInput{
		Engine:      aws.String("neptune"),
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

func testAccEngineVersionDataSourceConfig_basic(version string) string {
	return fmt.Sprintf(`
data "aws_neptune_engine_version" "test" {
  engine  = "neptune"
  version = %q
}
`, version)
}

func testAccEngineVersionDataSourceConfig_preferred() string {
	return `
data "aws_neptune_engine_version" "test" {
  preferred_versions = ["85.9.12", "1.0.3.0", "1.0.2.2"]
}
`
}

func testAccEngineVersionDataSourceConfig_defaultOnly() string {
	return `
data "aws_neptune_engine_version" "test" {}
`
}
