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
	rBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rObjectKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPluginName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigBasic(rBucketName, rObjectKey, rPluginName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision"),
					resource.TestCheckResourceAttr(resourceName, "state", kafkaconnect.CustomPluginStateActive),
					resource.TestCheckResourceAttr(resourceName, "content_type", kafkaconnect.CustomPluginContentTypeJar),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.bucket_arn", fmt.Sprintf("arn:aws:s3:::%s", rBucketName)),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.file_key", rObjectKey),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.object_version", ""),
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
	rBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rObjectKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPluginName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription := sdkacctest.RandString(20)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigDescription(rBucketName, rObjectKey, rPluginName, rDescription),
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
	rBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rObjectKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPluginName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPluginNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigBasic(rBucketName, rObjectKey, rPluginName),
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
				Config: testAccCustomPluginConfigContentTypeZip(rBucketName, rObjectKey, rPluginNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_type", kafkaconnect.CustomPluginContentTypeZip),
				),
			},
		},
	})
}

func TestAccKafkaConnectCustomPlugin_objectVersion(t *testing.T) {
	rBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rObjectKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPluginName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigObjectVersion(rBucketName, rObjectKey, rPluginName),
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
		return err
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

func testAccCustomPluginConfigBasicS3ObjectZip(bucketName string, objectKey string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[2]q
  source = "test-fixtures/activemq-connector.zip"
}
`, bucketName, objectKey)
}

func testAccCustomPluginConfigBasicS3ObjectJar(bucketName string, objectKey string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[2]q
  source = "test-fixtures/mongodb-connector.jar"
}
`, bucketName, objectKey)
}

func testAccCustomPluginConfigBasic(bucketName, objectKey, pluginName string) string {
	return acctest.ConfigCompose(testAccCustomPluginConfigBasicS3ObjectJar(bucketName, objectKey), fmt.Sprintf(`
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
`, pluginName))
}

func testAccCustomPluginConfigDescription(bucketName, objectKey, pluginName, description string) string {
	return acctest.ConfigCompose(testAccCustomPluginConfigBasicS3ObjectJar(bucketName, objectKey), fmt.Sprintf(`
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
`, pluginName, description))
}

func testAccCustomPluginConfigContentTypeZip(bucketName, objectKey, pluginName string) string {
	return acctest.ConfigCompose(testAccCustomPluginConfigBasicS3ObjectZip(bucketName, objectKey), fmt.Sprintf(`
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
`, pluginName))
}

func testAccCustomPluginConfigObjectVersion(bucketName, objectKey, pluginName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[2]q
  source = "test-fixtures/mongodb-connector.jar"
}

resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[3]q
  content_type = "JAR"
  location {
    s3 {
      bucket_arn     = aws_s3_bucket.test.arn
      file_key       = aws_s3_bucket_object.test.key
      object_version = aws_s3_bucket_object.test.version_id
    }
  }
}


`, bucketName, objectKey, pluginName)
}
