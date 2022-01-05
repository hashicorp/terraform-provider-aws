package kafkaconnect_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccKafkaConnectWorkerConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	propertiesFileContent := "key.converter=hello\nvalue.converter=world"

	resourceName := "aws_mskconnect_worker_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerConfigurationBasic(rName, propertiesFileContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkerConfigurationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "properties_file_content", propertiesFileContent),
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
	rDescription := sdkacctest.RandString(20)

	propertiesFileContent := "key.converter=hello\nvalue.converter=world"

	resourceName := "aws_mskconnect_worker_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerConfigurationDescription(rName, propertiesFileContent, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkerConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription),
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

func TestAccKafkaConnectWorkerConfiguration_properties_file_content(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	propertiesFileContent := "key.converter=hello\nvalue.converter=world"
	propertiesFileContentBase64 := base64.StdEncoding.EncodeToString([]byte(propertiesFileContent))

	resourceName := "aws_mskconnect_worker_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerConfigurationBasic(rName, propertiesFileContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkerConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "properties_file_content", propertiesFileContent),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkerConfigurationBasic(rName, propertiesFileContentBase64),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkerConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "properties_file_content", propertiesFileContent),
				),
			},
		},
	})
}

func testAccCheckWorkerConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Worker Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn

		params := &kafkaconnect.DescribeWorkerConfigurationInput{
			WorkerConfigurationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeWorkerConfiguration(params)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccWorkerConfigurationBasic(name, content string) string {
	return fmt.Sprintf(`
resource "aws_mskconnect_worker_configuration" "test" {
  name                    = %[1]q
  properties_file_content = %[2]q
}
`, name, content)
}

func testAccWorkerConfigurationDescription(name, content, description string) string {
	return fmt.Sprintf(`
resource "aws_mskconnect_worker_configuration" "test" {
  name                    = %[1]q
  properties_file_content = %[2]q
  description             = %[3]q
}
`, name, content, description)
}
