package kafka_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSMskKafkaVersionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_msk_kafka_version.test"
	version := "2.4.1.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccAWSMskKafkaVersionPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMskKafkaVersionDataSourceBasicConfig(version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", version),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
				),
			},
		},
	})
}

func TestAccAWSMskKafkaVersionDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_msk_kafka_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccAWSMskKafkaVersionPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMskKafkaVersionDataSourcePreferredConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", "2.4.1.1"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
				),
			},
		},
	})
}

func testAccAWSMskKafkaVersionPreCheck(t *testing.T) {
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

func testAccAWSMskKafkaVersionDataSourceBasicConfig(version string) string {
	return fmt.Sprintf(`
data "aws_msk_kafka_version" "test" {
  version = %[1]q
}
`, version)
}

func testAccAWSMskKafkaVersionDataSourcePreferredConfig() string {
	return `
data "aws_msk_kafka_version" "test" {
  preferred_versions = ["2.4.1.1", "2.4.1", "2.2.1"]
}
`
}
