package kafkaconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafkaconnect "github.com/hashicorp/terraform-provider-aws/internal/service/kafkaconnect"
)

func TestAccKafkaConnectWorkerConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_worker_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkerConfigurationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "properties_file_content", "key.converter=org.apache.kafka.connect.storage.StringConverter\nvalue.converter=org.apache.kafka.connect.storage.StringConverter\n"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKafkaConnectWorkerConfiguration_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_worker_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerConfigurationConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkerConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "testing"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckWorkerConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Connect Worker Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn

		_, err := tfkafkaconnect.FindWorkerConfigurationByARN(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccWorkerConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_mskconnect_worker_configuration" "test" {
  name = %[1]q

  properties_file_content = <<EOF
key.converter=org.apache.kafka.connect.storage.StringConverter
value.converter=org.apache.kafka.connect.storage.StringConverter
EOF
}
`, rName)
}

func testAccWorkerConfigurationConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_mskconnect_worker_configuration" "test" {
  name        = %[1]q
  description = "testing"

  properties_file_content = <<EOF
key.converter=org.apache.kafka.connect.storage.StringConverter
value.converter=org.apache.kafka.connect.storage.StringConverter
EOF
}
`, rName)
}
