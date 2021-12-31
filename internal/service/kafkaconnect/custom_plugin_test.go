package kafkaconnect_test

import (
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

func TestAccKafkaConnectCustomPlugin_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision"),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.s3.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.file_key", rName),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "state", kafkaconnect.CustomPluginStateActive),
					resource.TestCheckResourceAttr(resourceName, "content_type", kafkaconnect.CustomPluginContentTypeJar),
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

func TestAccKafkaConnectCustomPlugin_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription := sdkacctest.RandString(20)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigDescription(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
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

func TestAccKafkaConnectCustomPlugin_contentType(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_type", kafkaconnect.CustomPluginContentTypeJar),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomPluginConfigContentTypeZip(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_type", kafkaconnect.CustomPluginContentTypeZip),
				),
			},
		},
	})
}

func TestAccKafkaConnectCustomPlugin_objectVersion(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigObjectVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					testAccCheckCustomPluginObjectVersion(resourceName),
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

func testAccCheckCustomPluginExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Custom Plugin ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn

		_, err := tfkafkaconnect.FindCustomPluginByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckCustomPluginObjectVersion(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		plugin, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type == "aws_s3_bucket_object" {
				pluginObjectVersion := plugin.Primary.Attributes["location.0.s3.0.object_version"]
				objectVersionId := rs.Primary.Attributes["version_id"]

				if !(pluginObjectVersion == objectVersionId) {
					return fmt.Errorf("Plugin object version doesn't match object's version id: %s != %s", pluginObjectVersion, objectVersionId)
				}

				return nil
			}
		}

		return fmt.Errorf("Couldn't find aws_s3_bucket_object resource to compare versions.")
	}
}

func testAccCustomPluginConfigBasicS3ObjectZip(name string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[1]q
  source = "test-fixtures/activemq-connector.zip"
}
`, name)
}

func testAccCustomPluginConfigBasicS3ObjectJar(name string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[1]q
  source = "test-fixtures/mongodb-connector.jar"
}
`, name)
}

func testAccCustomPluginConfigBasic(name string) string {
	return acctest.ConfigCompose(testAccCustomPluginConfigBasicS3ObjectJar(name), fmt.Sprintf(`
resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[1]q
  content_type = "JAR"

  location {
    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = aws_s3_bucket_object.test.key
    }
  }
}
`, name))
}

func testAccCustomPluginConfigDescription(name, description string) string {
	return acctest.ConfigCompose(testAccCustomPluginConfigBasicS3ObjectJar(name), fmt.Sprintf(`
resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[1]q
  description  = %[2]q
  content_type = "JAR"

  location {
    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = aws_s3_bucket_object.test.key
    }
  }
}
`, name, description))
}

func testAccCustomPluginConfigContentTypeZip(name string) string {
	return acctest.ConfigCompose(testAccCustomPluginConfigBasicS3ObjectZip(name), fmt.Sprintf(`
resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[1]q
  content_type = "ZIP"

  location {
    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = aws_s3_bucket_object.test.key
    }
  }
}
`, name))
}

func testAccCustomPluginConfigObjectVersion(name string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[1]q
  source = "test-fixtures/mongodb-connector.jar"
}

resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[1]q
  content_type = "JAR"

  location {
    s3 {
      bucket_arn     = aws_s3_bucket.test.arn
      file_key       = aws_s3_bucket_object.test.key
      object_version = aws_s3_bucket_object.test.version_id
    }
  }
}
`, name)
}
