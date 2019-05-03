package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDBEngineVersionDataSource_mariadb(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDBEngineVersionDataSourceConfig("mariadb"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "db_engine_description", "MariaDB Community Edition"),
					resource.TestMatchResourceAttr("data.aws_db_engine_version.test", "db_engine_version_description", regexp.MustCompile("^MariaDB ")),
					resource.TestMatchResourceAttr("data.aws_db_engine_version.test", "db_parameter_group_family", regexp.MustCompile("^mariadb")),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "engine", "mariadb"),
					resource.TestMatchResourceAttr("data.aws_db_engine_version.test", "engine_version", regexp.MustCompile(`^\d+\.\d+\.\d+$`)),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "exportable_log_types.#", "4"),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "supported_engine_modes.#", "0"),
					resource.TestCheckResourceAttrSet("data.aws_db_engine_version.test", "supported_feature_names.#"),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "supports_log_exports_to_cloudwatch_logs", "true"),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "supports_read_replica", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBEngineVersionDataSource_postgres(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDBEngineVersionDataSourceConfig("postgres"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "db_engine_description", "PostgreSQL"),
					resource.TestMatchResourceAttr("data.aws_db_engine_version.test", "db_engine_version_description", regexp.MustCompile("^PostgreSQL ")),
					resource.TestMatchResourceAttr("data.aws_db_engine_version.test", "db_parameter_group_family", regexp.MustCompile("^postgres")),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "engine", "postgres"),
					resource.TestMatchResourceAttr("data.aws_db_engine_version.test", "engine_version", regexp.MustCompile(`^\d+\.\d+$`)),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "exportable_log_types.#", "2"),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "supported_engine_modes.#", "0"),
					resource.TestCheckResourceAttrSet("data.aws_db_engine_version.test", "supported_feature_names.#"),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "supports_log_exports_to_cloudwatch_logs", "true"),
					resource.TestCheckResourceAttr("data.aws_db_engine_version.test", "supports_read_replica", "true"),
				),
			},
		},
	})
}

func testAccCheckAwsDBEngineVersionDataSourceConfig(engine string) string {
	return fmt.Sprintf(`
data "aws_db_engine_version" "test" {
  engine      = "%s"
  most_recent = true
}
`, engine)
}
