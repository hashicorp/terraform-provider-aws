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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccKafkaConnectCustomPlugin_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      testAccCheckCustomPluginDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "ZIP"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.s3.0.bucket_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.s3.0.file_key"),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
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

func TestAccKafkaConnectCustomPlugin_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      testAccCheckCustomPluginDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfkafkaconnect.ResourceCustomPlugin(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaConnectCustomPlugin_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      testAccCheckCustomPluginDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
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

func TestAccKafkaConnectCustomPlugin_objectVersion(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      testAccCheckCustomPluginDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfigObjectVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.s3.0.object_version"),
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
			return fmt.Errorf("No MSK Connect Custom Plugin ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn

		_, err := tfkafkaconnect.FindCustomPluginByARN(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckCustomPluginDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_mskconnect_custom_plugin" {
			continue
		}

		_, err := tfkafkaconnect.FindCustomPluginByARN(context.TODO(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("MSK Connect Custom Plugin %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCustomPluginBaseConfig(rName string, s3BucketVersioning bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket_acl.test.bucket

  versioning_configuration {
    status = %[2]t ? "Enabled" : "Suspended"
  }
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket_versioning.test.bucket
  key    = "jcustenborder-kafka-connect-simulator-0.1.120.zip"
  source = "test-fixtures/jcustenborder-kafka-connect-simulator-0.1.120.zip"
}
`, rName, s3BucketVersioning)
}

func testAccCustomPluginConfig(rName string) string {
	return acctest.ConfigCompose(testAccCustomPluginBaseConfig(rName, false), fmt.Sprintf(`
resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[1]q
  content_type = "ZIP"

  location {
    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = aws_s3_object.test.key
    }
  }
}
`, rName))
}

func testAccCustomPluginConfigDescription(rName string) string {
	return acctest.ConfigCompose(testAccCustomPluginBaseConfig(rName, false), fmt.Sprintf(`
resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[1]q
  content_type = "ZIP"
  description  = "testing"

  location {
    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = aws_s3_object.test.key
    }
  }
}
`, rName))
}

func testAccCustomPluginConfigObjectVersion(rName string) string {
	return acctest.ConfigCompose(testAccCustomPluginBaseConfig(rName, true), fmt.Sprintf(`
resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[1]q
  content_type = "ZIP"
  description  = "testing"

  location {
    s3 {
      bucket_arn     = aws_s3_bucket.test.arn
      file_key       = aws_s3_object.test.key
      object_version = aws_s3_object.test.version_id
    }
  }
}
`, rName))
}
