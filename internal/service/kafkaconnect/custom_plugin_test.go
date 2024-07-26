// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafkaconnect "github.com/hashicorp/terraform-provider-aws/internal/service/kafkaconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaConnectCustomPlugin_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckCustomPluginDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPluginExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "ZIP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision"),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.s3.0.bucket_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.s3.0.file_key"),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ACTIVE"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckCustomPluginDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafkaconnect.ResourceCustomPlugin(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaConnectCustomPlugin_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckCustomPluginDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
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

func TestAccKafkaConnectCustomPlugin_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckCustomPluginDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPluginExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomPluginConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPluginExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCustomPluginConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPluginExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKafkaConnectCustomPlugin_objectVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckCustomPluginDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginConfig_objectVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomPluginExists(ctx, resourceName),
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

func testAccCheckCustomPluginExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectClient(ctx)

		_, err := tfkafkaconnect.FindCustomPluginByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckCustomPluginDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mskconnect_custom_plugin" {
				continue
			}

			_, err := tfkafkaconnect.FindCustomPluginByARN(ctx, conn, rs.Primary.ID)

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
}

func testAccCustomPluginBaseConfig(rName string, s3BucketVersioning bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket

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

func testAccCustomPluginConfig_basic(rName string) string {
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

func testAccCustomPluginConfig_description(rName string) string {
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

func testAccCustomPluginConfig_objectVersion(rName string) string {
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

func testAccCustomPluginConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccCustomPluginConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
