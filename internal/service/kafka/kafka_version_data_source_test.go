package kafka_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccKafkaKafkaVersionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_msk_kafka_version.test"
	version := "2.4.1.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccVersionBasicDataSourceConfig(version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", version),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
				),
			},
		},
	})
}

func TestAccKafkaKafkaVersionDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_msk_kafka_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccVersionPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccVersionPreferredDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", "2.4.1.1"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
				),
			},
		},
	})
}

func testAccVersionPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConn

	input := &kafka.ListKafkaVersionsInput{}

	_, err := conn.ListKafkaVersions(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVersionBasicDataSourceConfig(version string) string {
	return fmt.Sprintf(`
data "aws_msk_kafka_version" "test" {
  version = %[1]q
}
`, version)
}

func testAccVersionPreferredDataSourceConfig() string {
	return `
data "aws_msk_kafka_version" "test" {
  preferred_versions = ["2.4.1.1", "2.4.1", "2.2.1"]
}
`
}
