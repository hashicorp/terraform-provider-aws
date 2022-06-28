package docdb_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDocDBEngineVersionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_docdb_engine_version.test"
	engine := "docdb"
	version := "3.6.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_basic(engine, version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "version", version),

					resource.TestCheckResourceAttrSet(dataSourceName, "engine_description"),
					resource.TestMatchResourceAttr(dataSourceName, "exportable_log_types.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "parameter_group_family"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_log_exports_to_cloudwatch"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_description"),
				),
			},
		},
	})
}

func TestAccDocDBEngineVersionDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_docdb_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferred(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", "3.6.0"),
				),
			},
		},
	})
}

func TestAccDocDBEngineVersionDataSource_defaultOnly(t *testing.T) {
	dataSourceName := "data.aws_docdb_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_defaultOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", "docdb"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version"),
				),
			},
		},
	})
}

func testAccEngineVersionPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

	input := &docdb.DescribeDBEngineVersionsInput{
		Engine:      aws.String("docdb"),
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

func testAccEngineVersionDataSourceConfig_basic(engine, version string) string {
	return fmt.Sprintf(`
data "aws_docdb_engine_version" "test" {
  engine  = %q
  version = %q
}
`, engine, version)
}

func testAccEngineVersionDataSourceConfig_preferred() string {
	return `
data "aws_docdb_engine_version" "test" {
  preferred_versions = ["34.6.1", "3.6.0", "2.6.0"]
}
`
}

func testAccEngineVersionDataSourceConfig_defaultOnly() string {
	return `
data "aws_docdb_engine_version" "test" {}
`
}
